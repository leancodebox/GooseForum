package component

import (
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"net/http"
)

type Status int

const (
	SUCCESS Status = iota // 成功
	FAIL                  // 失败
)

type BetterRequest[T any] struct {
	Params   T
	UserId   uint64
	userSet  bool
	userInfo users.EntityComplete
}
type Null struct {
}
type NullRequest BetterRequest[Null]

func (r *BetterRequest[T]) GetParams() T {
	return r.Params
}

func (r *BetterRequest[T]) GetUser() (users.EntityComplete, error) {
	if r.userSet != false {
		return r.userInfo, nil
	}
	user, _ := users.Get(r.UserId)
	r.userSet = true
	r.userInfo = user
	return r.userInfo, nil
}

type Response struct {
	Code int
	Data ResultStruct
}

type ResultStruct struct {
	Msg    any    `json:"msg"`
	Result any    `json:"result"`
	Code   Status `json:"code"`
}

type DataMap map[string]interface{}

func BuildResponse(code int, data ResultStruct) Response {
	return Response{code, data}
}

func SuccessResponse(data any) Response {
	return BuildResponse(http.StatusOK,
		SuccessData(data),
	)
}

func SuccessPage[T any](list []T, page, size int, total int64) Response {
	return SuccessResponse(Page[T]{List: list, Page: page, Total: total, Size: size})
}

func SuccessData(data any) ResultStruct {
	return ResultStruct{
		Msg:    nil,
		Result: data,
		Code:   SUCCESS,
	}
}

func FailResponse(msg any) Response {
	return BuildResponse(http.StatusOK,
		FailData(msg),
	)
}

func FailData(msg any) ResultStruct {
	return ResultStruct{
		Msg:    msg,
		Result: nil,
		Code:   FAIL,
	}
}
