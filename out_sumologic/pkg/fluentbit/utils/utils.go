package utils

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/fluent/fluent-bit-go/output"
)

func encodeJSON(record map[interface{}]interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	for k, v := range record {
		switch t := v.(type) {
		case []byte:
			// prevent encoding to base64
			m[k.(string)] = string(t)
		case map[interface{}]interface{}:
			if nextValue, ok := record[k].(map[interface{}]interface{}); ok {
				m[k.(string)] = encodeJSON(nextValue)
			}
		default:
			m[k.(string)] = v
		}
	}
	return m
}

func flatten(m map[string]interface{}) map[string]interface{} {
	o := make(map[string]interface{})
	for k, v := range m {
		switch child := v.(type) {
		case map[string]interface{}:
			nm := flatten(child)
			for nk, nv := range nm {
				o[k+"."+nk] = nv
			}
		default:
			o[k] = v
		}
	}
	return o
}

func CreateJSON(record map[interface{}]interface{}, logKey string) (string, error) {
	var res []byte
	var err error

	m := flatten(encodeJSON(record))
	log, ok := m[logKey]

	if ok {
		res, err = json.Marshal(log)
	} else {
		res, err = json.Marshal(m)
	}
	if err != nil {
		return string("{}"), err
	}
	return strconv.Unquote(string(res))
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
