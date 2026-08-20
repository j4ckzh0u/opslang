// Package type_debug provides Ansible type_debug module equivalent.
// Returns the type information of a value.
package type_debug

import (
	"fmt"
	"reflect"
)

// TypeResult is returned by Debug.
type TypeResult struct {
	Value      interface{} `json:"value"`
	Type       string      `json:"type"`
	GoType     string      `json:"go_type"`
	IsList     bool        `json:"is_list"`
	IsDict     bool        `json:"is_dict"`
	IsString   bool        `json:"is_string"`
	IsNumber   bool        `json:"is_number"`
	IsBool     bool        `json:"is_bool"`
	IsNull     bool        `json:"is_null"`
	Length     int         `json:"length,omitempty"`
	Keys       []string    `json:"keys,omitempty"`
	Error      string      `json:"error,omitempty"`
}

// Debug returns type information about a value.
func Debug(value interface{}) TypeResult {
	if value == nil {
		return TypeResult{Type: "null", GoType: "nil", IsNull: true}
	}

	r := TypeResult{
		Value:  value,
		GoType: reflect.TypeOf(value).String(),
	}

	switch v := value.(type) {
	case string:
		r.Type = "string"
		r.IsString = true
		r.Length = len(v)
	case bool:
		r.Type = "bool"
		r.IsBool = true
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		r.Type = "number"
		r.IsNumber = true
	case []interface{}:
		r.Type = "list"
		r.IsList = true
		r.Length = len(v)
	case map[string]interface{}:
		r.Type = "dict"
		r.IsDict = true
		r.Length = len(v)
		keys := make([]string, 0, len(v))
		for k := range v {
			keys = append(keys, k)
		}
		r.Keys = keys
	default:
		rv := reflect.ValueOf(value)
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			r.Type = "list"
			r.IsList = true
			r.Length = rv.Len()
		case reflect.Map:
			r.Type = "dict"
			r.IsDict = true
			r.Length = rv.Len()
			keys := make([]string, 0, rv.Len())
			for _, k := range rv.MapKeys() {
				keys = append(keys, fmt.Sprintf("%v", k.Interface()))
			}
			r.Keys = keys
		case reflect.String:
			r.Type = "string"
			r.IsString = true
			r.Length = rv.Len()
		case reflect.Bool:
			r.Type = "bool"
			r.IsBool = true
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			r.Type = "number"
			r.IsNumber = true
		default:
			r.Type = "unknown"
		}
	}
	return r
}
