// Package jsonopt provides compact JSON encode/decode helpers.
package jsonopt

import (
	"encoding/json"
)

// EncodeE marshals obj into a compact JSON string.
func EncodeE(obj any) (string, error) {
	marshal, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}

// Encode marshals obj into JSON and ignores marshal errors.
func Encode(obj any) string {
	str, _ := EncodeE(obj)
	return str
}

// EncodeFormatE marshals obj into an indented JSON string.
func EncodeFormatE(obj any) (string, error) {
	marshal, err := json.MarshalIndent(obj, "", " ")
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}

// EncodeFormat marshals obj into indented JSON and ignores marshal errors.
func EncodeFormat(obj any) string {
	str, _ := EncodeFormatE(obj)
	return str
}

// DecodeE unmarshals JSON from str into T.
func DecodeE[T any, P string | []byte](str P) (T, error) {
	var obj T
	err := json.Unmarshal([]byte(str), &obj)
	return obj, err
}

// Decode unmarshals JSON from str into T and ignores unmarshal errors.
func Decode[T any, P string | []byte](str P) T {
	entity, _ := DecodeE[T](str)
	return entity
}
