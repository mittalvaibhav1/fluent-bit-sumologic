package main

import (
	"C"
	"out_sumologic/pkg/fluentbit/logger"
	"out_sumologic/pkg/sumologic"
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
	"github.com/sirupsen/logrus"
)

var (
	instances []*sumologic.SumoLogic
	log       *logrus.Entry
)

//export FLBPluginRegister
func FLBPluginRegister(def unsafe.Pointer) int {
	// Gets called only once when the plugin.so is loaded
	return output.FLBPluginRegister(def, sumologic.PLUGIN_NAME, sumologic.PLUGIN_DESCRIPTION)
}

//export FLBPluginInit
func FLBPluginInit(plugin unsafe.Pointer) int {
	// Gets called only once for each instance you have configured
	id := len(instances)

	// Initalize a sumologic instance
	instance, err := sumologic.Initalize(plugin, id)
	if err != nil {
		log.Fatal("unable to initalise the plugin", err)
		return output.FLB_ERROR
	}

	instances = append(instances, instance)
	output.FLBPluginSetContext(plugin, id)
	return output.FLB_OK
}

//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
	log.Warn("flush called for unknown instance")
	return output.FLB_OK
}

//export FLBPluginFlushCtx
func FLBPluginFlushCtx(ctx, data unsafe.Pointer, length C.int, tag *C.char) int {
	// Gets called with a batch of records to be written to an instance
	instance := instances[output.FLBPluginGetContext(ctx).(int)]

	// Create a batch from the records
	batch, err := instance.CreateBatch(data, int(length), C.GoString(tag))
	if err != nil {
		return output.FLB_RETRY
	}

	// Send the batch to sumologic
	err = instance.SendBatch(batch)
	if err != nil {
		return output.FLB_RETRY
	}

	return output.FLB_OK
}

//export FLBPluginExit
func FLBPluginExit() int {
	log.Warn("exit called for unknown instance")
	return output.FLB_OK
}

//export FLBPluginExitCtx
func FLBPluginExitCtx(ctx unsafe.Pointer) int {
	instance := instances[output.FLBPluginGetContext(ctx).(int)]
	instance.Stop()
	return output.FLB_OK
}

//export FLBPluginUnregister
func FLBPluginUnregister(def unsafe.Pointer) {
	output.FLBPluginUnregister(def)
}

func init() {
	// FIXME(mittalvaibhav1, 09-03-2024): should be replaced with the service log level once that is supported for Go plugins.
	log = logger.New(sumologic.PLUGIN_NAME, logrus.DebugLevel)
}

func main() {}
