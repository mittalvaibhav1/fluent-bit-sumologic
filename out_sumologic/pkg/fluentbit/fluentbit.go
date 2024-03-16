package fluentbit

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
)

type FLBPluginConfig struct {
	Plugin unsafe.Pointer
}

func (c *FLBPluginConfig) Get(key string) string {
	return output.FLBPluginConfigKey(c.Plugin, key)
}

func (c *FLBPluginConfig) GetOrDefault(key string, defaultValue string) string {
	value := c.Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetOutputInstanceName(name string, id int) string {
	// Similar pattern is followed by all fluent-bit plugins
	return fmt.Sprintf("output:%s:%s.%d", name, name, id)
}

func GetTimeStamp(ts interface{}) time.Time {
	switch t := ts.(type) {
	case output.FLBTime:
		return ts.(output.FLBTime).Time
	case uint64:
		return time.Unix(int64(t), 0)
	default:
		return time.Now()
	}
}
