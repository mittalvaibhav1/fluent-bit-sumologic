package sumologic

import (
	"fmt"
	"os"
	"out_sumologic/pkg/fluentbit"
	"regexp"
	"strconv"
	"strings"
	"unsafe"

	"github.com/sirupsen/logrus"
)

const (
	DEFAULT_SOURCE_CATEGORY = "sumologic_default"
	DEFAULT_TAG_DELIMITER   = "."
	DEFAULT_LOG_LEVEL       = "info"
	DEFAULT_MAX_RETRIES     = "10"
	DEFAULT_LOG_KEY         = "log"
	TAG_REGEX               = `\$TAG\[(\d+)\]`
)

type sumoLogicConfig struct {
	collectorURL   string // Required
	sourceName     string
	sourceHost     string
	sourceCategory string
	tagDelimiter   string
	level          logrus.Level
	logKey         string
	maxRetries     uint64
}

func (c *sumoLogicConfig) String() string {
	return fmt.Sprintf("Collector_Url=%s, Source_Name=%s, Source_Host=%s, Source_Category=%s, Tag_Delimiter=%s, Level=%s, Log_Key=%s, Max_Retries=%d",
		c.collectorURL, c.sourceName, c.sourceHost, c.sourceCategory, c.tagDelimiter, c.level.String(), c.logKey, c.maxRetries)
}

func loadConfig(plugin unsafe.Pointer) (*sumoLogicConfig, error) {
	var err error
	// Initalize and load config
	config := new(sumoLogicConfig)
	c := fluentbit.FLBPluginConfig{
		Plugin: plugin,
	}

	config.collectorURL = c.Get("Collector_Url")
	if config.collectorURL == "" {
		return nil, fmt.Errorf("invalid value for Collector_Url, it cannot be empty")
	}

	sourceHost, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get the hostname, %w", err)
	}
	config.sourceHost = c.GetOrDefault("Source_Host", sourceHost)

	config.sourceName = c.Get("Source_Name")
	config.sourceCategory = c.GetOrDefault("Source_Category", DEFAULT_SOURCE_CATEGORY)
	config.tagDelimiter = c.GetOrDefault("Tag_Delimiter", DEFAULT_TAG_DELIMITER)
	config.logKey = c.GetOrDefault("Log_Key", DEFAULT_LOG_KEY)

	config.level, err = logrus.ParseLevel(c.GetOrDefault("Level", DEFAULT_LOG_LEVEL))
	if err != nil {
		return nil, fmt.Errorf("invalid value for Level, %w", err)
	}

	config.maxRetries, err = strconv.ParseUint(c.GetOrDefault("Max_Retries", DEFAULT_MAX_RETRIES), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid value for Retry_Limit, %w", err)
	}

	return config, nil
}

func replaceWithTag(value string, tagSlice []string) (string, error) {
	regex := regexp.MustCompile(TAG_REGEX)
	length := len(tagSlice)
	matches := regex.FindAllString(value, length)

	for _, match := range matches {
		stringIndex := regex.FindStringSubmatch(match)[1]
		index, err := strconv.Atoi(stringIndex)
		if err != nil {
			return "", err
		}
		if index >= length {
			return "", fmt.Errorf("tag index %d out of bounds. length %d", index, length)
		}
		value = strings.Replace(value, match, tagSlice[index], 1)
	}

	return value, nil
}

func parseConfig(config *sumoLogicConfig, tag string) (*sumoLogicConfig, error) {
	var err error
	tagSlice := strings.Split(tag, config.tagDelimiter)

	batchConfig := *config
	batchConfig.sourceName, err = replaceWithTag(batchConfig.sourceName, tagSlice)
	if err != nil {
		return nil, err
	}
	batchConfig.sourceCategory, err = replaceWithTag(batchConfig.sourceCategory, tagSlice)
	if err != nil {
		return nil, err
	}
	batchConfig.sourceHost, err = replaceWithTag(batchConfig.sourceHost, tagSlice)
	if err != nil {
		return nil, err
	}

	return &batchConfig, nil
}
