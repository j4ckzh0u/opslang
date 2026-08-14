// Package opsjson provides JSON encoding and decoding operations for OpsLang.
// All functions return structured results with JSON tags, enabling easy
// serialization and downstream processing. Uses pure Go stdlib encoding/json.
package opsjson

import (
	"encoding/json"
	"fmt"
)

// EncodeResult is returned by Encode, holding the JSON string and its size in bytes.
type EncodeResult struct {
	JSON string `json:"json"`
	Size int    `json:"size"`
}

// DecodeResult is returned by Decode, holding the deserialized data.
type DecodeResult struct {
	Data interface{} `json:"data"`
}

// Encode marshals data into an indented JSON string using 2-space indent.
func Encode(data interface{}) (EncodeResult, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return EncodeResult{}, fmt.Errorf("opsjson.Encode: %w", err)
	}
	return EncodeResult{
		JSON: string(bytes),
		Size: len(bytes),
	}, nil
}

// Decode unmarshals a JSON string into an interface{} value.
// Numbers are decoded as float64, objects as map[string]interface{},
// arrays as []interface{}, and so on per encoding/json defaults.
func Decode(input string) (DecodeResult, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return DecodeResult{}, fmt.Errorf("opsjson.Decode: %w", err)
	}
	return DecodeResult{Data: data}, nil
}
