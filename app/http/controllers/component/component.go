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
	userInfo users.Entity
}
type Null struct {
}
type NullRequest BetterRequest[Null]

func (r *BetterRequest[T]) GetParams() T {
	return r.Params
}

func (r *BetterRequest[T]) GetUser() (users.Entity, error) {
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
	Data any
}

type DataMap map[string]interface{}

func BuildResponse(code int, data any) Response {
	return Response{code, data}
}

func SuccessResponse(data any) Response {
	return BuildResponse(http.StatusOK,
		SuccessData(data),
	)
}

func SuccessPage[T any](list []T, size int, total int64) Response {
	return SuccessResponse(Page[T]{List: list, Total: total, Size: size})
}

func SuccessData(data any) map[string]any {
	return map[string]any{
		"msg":    nil,
		"result": data,
		"code":   SUCCESS,
	}
}

func FailResponse(msg any) Response {
	return BuildResponse(http.StatusOK,
		FailData(msg),
	)
}

func FailData(msg any) map[string]any {
	return map[string]any{
		"msg":    msg,
		"result": nil,
		"code":   FAIL,
	}
}
