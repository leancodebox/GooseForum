package component

import (
	"github.com/leancodebox/GooseForum/app/models/Users"
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
	userInfo Users.Users
}
type Null struct {
}
type NullRequest BetterRequest[Null]

func (r *BetterRequest[T]) GetParams() T {
	return r.Params
}

func (r *BetterRequest[T]) GetUser() (Users.Users, error) {
	if r.userSet != false {
		return r.userInfo, nil
	}
	user, _ := Users.Get(r.UserId)
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
		map[string]any{
			"msg":    nil,
			"result": data,
			"code":   SUCCESS,
		},
	)
}

func FailResponse(msg any) Response {
	return BuildResponse(http.StatusOK,
		map[string]any{
			"msg":    msg,
			"result": nil,
			"code":   FAIL,
		},
	)
}
