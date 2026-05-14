package jsonopt

import "testing"

func TestJsonEncode(t *testing.T) {
	type tmp struct {
		Name string
	}
	if got := Encode(tmp{Name: "name"}); got != `{"Name":"name"}` {
		t.Fatalf("Encode() = %s", got)
	}
}

func TestJsonDecode(t *testing.T) {
	type tmp struct {
		Name string
	}
	if got := Decode[tmp](`{"name":"name"}`); got.Name != "name" {
		t.Fatalf("Decode().Name = %q", got.Name)
	}
	type Cat struct {
		Id int `json:"id"`
	}
	type DogKing[T any] struct {
		Body T `json:"body"`
	}
	catList, _ := DecodeE[[]Cat](`[{"id":1}]`)
	if len(catList) != 1 || catList[0].Id != 1 {
		t.Fatalf("DecodeE list = %#v", catList)
	}
	catMap, _ := DecodeE[map[string]Cat](`{"ok":{"id":1}}`)
	if catMap["ok"].Id != 1 {
		t.Fatalf("DecodeE map = %#v", catMap)
	}
	dk, _ := DecodeE[DogKing[Cat]](`{"body":{"id":1231}}`)
	if dk.Body.Id != 1231 {
		t.Fatalf("DecodeE generic struct = %#v", dk)
	}
	type HighDogKing map[string]DogKing[Cat]
	dk2, _ := DecodeE[HighDogKing](`{"key":{"body":{"id":1231}}}`)
	if dk2["key"].Body.Id != 1231 {
		t.Fatalf("DecodeE generic map = %#v", dk2)
	}
}
