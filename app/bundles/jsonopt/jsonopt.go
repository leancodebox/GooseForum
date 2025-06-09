package jsonopt

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func init() {
	extra.RegisterFuzzyDecoders()
}
func EncodeE(obj any) (string, error) {
	marshal, err := json.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}

func Encode(obj any) string {
	str, _ := EncodeE(obj)
	return str
}

func EncodeFormatE(obj any) (string, error) {
	marshal, err := json.MarshalIndent(obj, "", " ")
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}

func EncodeFormat(obj any) string {
	str, _ := EncodeFormatE(obj)
	return str
}

func DecodeE[T any, P string | []byte](str P) (T, error) {
	var obj T
	err := json.Unmarshal([]byte(str), &obj)
	return obj, err
}

func Decode[T any, P string | []byte](str P) T {
	entity, _ := DecodeE[T](str)
	return entity
}
