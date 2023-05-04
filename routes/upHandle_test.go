package routes

import (
	"fmt"
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
		fmt.Println(m0.In(i))
		data := reflect.New(m0.In(0))
		fmt.Println(data)
	}
	return func(s string) string {
		return s
	}
}

func TestName(t *testing.T) {
	upCatAction(CatAction)
}
