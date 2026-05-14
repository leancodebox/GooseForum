package routes

import (
	"reflect"
	"testing"
)

type CatActionReq struct {
	Id uint64 `json:"id"`
}

type Response struct {
	data string
}

func CatAction(req CatActionReq) Response {
	return Response{data: "131231"}
}

func upCatAction(params any) func(string) string {
	m0 := reflect.TypeOf(params)
	for i := 0; i < m0.NumIn(); i++ {
		_ = reflect.New(m0.In(i))
	}
	return func(s string) string {
		return s
	}
}

func TestName(t *testing.T) {
	upCatAction(CatAction)
}
