// Package opsyaml provides YAML encoding and decoding operations for OpsLang.
// All functions return structured results with JSON tags, enabling easy
// serialization and downstream processing. Uses gopkg.in/yaml.v3.
package opsyaml

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// EncodeResult is returned by Encode, holding the YAML string and its size in bytes.
type EncodeResult struct {
	YAML string `json:"yaml"`
	Size int    `json:"size"`
}

// DecodeResult is returned by Decode, holding the deserialized data.
type DecodeResult struct {
	Data interface{} `json:"data"`
}

// Encode marshals data into a YAML string.
func Encode(data interface{}) (EncodeResult, error) {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return EncodeResult{}, fmt.Errorf("opsyaml.Encode: %w", err)
	}
	return EncodeResult{
		YAML: string(bytes),
		Size: len(bytes),
	}, nil
}

// Decode unmarshals a YAML string into an interface{} value.
// Maps are decoded as map[string]interface{}, sequences as []interface{},
// and scalar values as their natural Go types.
func Decode(input string) (DecodeResult, error) {
	var data interface{}
	if err := yaml.Unmarshal([]byte(input), &data); err != nil {
		return DecodeResult{}, fmt.Errorf("opsyaml.Decode: %w", err)
	}
	return DecodeResult{Data: data}, nil
}
