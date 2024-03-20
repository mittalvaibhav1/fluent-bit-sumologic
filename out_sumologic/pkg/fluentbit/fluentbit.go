package fluentbit

import (
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
)

type FLBPluginConfig struct {
	plugin unsafe.Pointer
}

func (c *FLBPluginConfig) Get(key string) string {
	return output.FLBPluginConfigKey(c.plugin, key)
}

func (c *FLBPluginConfig) GetOrDefault(key string, defaultValue string) string {
	value := c.Get(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func New(plugin unsafe.Pointer) *FLBPluginConfig {
	return &FLBPluginConfig{
		plugin: plugin,
	}
}
