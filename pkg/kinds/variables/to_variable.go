// Package json @Author larry
// @Date 2025/4/11 11:07
// @Desc

package variables

import (
	"fmt"
	"reflect"
	"strings"
)

// meta.active: true
// age: 30
// address.zipcode: 10001
// address.locations[0].longitude: -74.006
// address.locations[1].longitude: -118.2437
// tags[0]: golang
// meta.additional: info
// meta.score: 99.5
// name: John Doe
// address.city: New York
// address.locations[0].latitude: 40.7128
// address.locations[1].latitude: 34.0522
// tags[1]: programming

// ToVariables converts structs, maps, or arrays to a flat map with JSONPath-like keys
func ToVariables(v interface{}) Variables {
	return collectVariables(reflect.ValueOf(v), "")
}

// mergeWithPrefix adjusts keys with a prefix and returns a new Variables map
func mergeWithPrefix(prefix string, variable Variables) Variables {
	newVariables := make(Variables)
	for k, v := range variable {
		if prefix != "" && k != "" {
			if k[0] == '[' {
				k = prefix + k // Do not add a dot if the next part is an array index
			} else {
				k = prefix + "." + k
			}
		} else if prefix != "" {
			k = prefix
		}
		newVariables[k] = v
	}
	return newVariables
}

// collectVariables recursively collects variables and returns as a Variables map
func collectVariables(v reflect.Value, prefix string) Variables {
	variables := make(Variables)

	switch v.Kind() {
	case reflect.Ptr:
		if !v.IsNil() {
			return collectVariables(v.Elem(), prefix)
		}
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			fieldValue := v.Field(i)
			fieldName := strings.ToLower(field.Name[:1]) + field.Name[1:] // Parse first letter to lowercase
			subVariables := collectVariables(fieldValue, fieldName)
			mergedVariables := mergeWithPrefix(prefix, subVariables)
			for k, val := range mergedVariables {
				variables[k] = val
			}
		}
	case reflect.Map:
		if prefix == "" {
			prefix = "map"
		}
		for _, key := range v.MapKeys() {
			val := v.MapIndex(key)
			mapKey := fmt.Sprintf("%v", key.Interface())
			subVariables := collectVariables(val, mapKey)
			mergedVariables := mergeWithPrefix(prefix, subVariables)
			for k, val := range mergedVariables {
				variables[k] = val
			}
		}
	case reflect.Array, reflect.Slice:
		if prefix == "" {
			prefix = "array"
		}
		if v.Len() == 0 {
			return variables
		}

		if v.Index(0).Kind() != reflect.Struct {
			variables[prefix] = fmt.Sprintf("%v", v.Interface())
			return variables
		}
		for i := 0; i < v.Len(); i++ {
			indexKey := fmt.Sprintf("[%d]", i)
			subVariables := collectVariables(v.Index(i), indexKey)
			mergedVariables := mergeWithPrefix(prefix, subVariables)
			for k, val := range mergedVariables {
				variables[k] = val
			}
		}
	default:
		if v.IsValid() {
			variables[prefix] = fmt.Sprintf("%v", v.Interface())
		}
	}

	return variables
}
