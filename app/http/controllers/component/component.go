package component

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
)

type Status int

const (
	SUCCESS Status = iota // 成功
	FAIL                  // 失败
)

type BetterRequest[T any] struct {
	Params     T
	UserId     uint64
	userSet    bool
	userInfo   users.EntityComplete
	GinContext *gin.Context
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
	Result      any           `json:"result"`
	Code        Status        `json:"code"`
	MessageCode MessageCode   `json:"messageCode,omitempty"`
	Params      MessageParams `json:"params,omitempty"`
}

type DataMap map[string]any

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
		Result: data,
		Code:   SUCCESS,
	}
}

// SuccessDataCode returns a successful response with a stable message code.
func SuccessDataCode(data any, messageCode MessageCode, params MessageParams) ResultStruct {
	return ResultStruct{
		Result:      data,
		Code:        SUCCESS,
		MessageCode: messageCode,
		Params:      params,
	}
}

func SuccessResponseCode(data any, messageCode MessageCode, params MessageParams) Response {
	return BuildResponse(http.StatusOK, SuccessDataCode(data, messageCode, params))
}

func FailResponse() Response {
	return BuildResponse(http.StatusOK,
		FailData(),
	)
}

func FailData() ResultStruct {
	return ResultStruct{
		Result: nil,
		Code:   FAIL,
	}
}

// FailDataCode returns a failed response with a stable message code.
func FailDataCode(messageCode MessageCode, params MessageParams) ResultStruct {
	return ResultStruct{
		Result:      nil,
		Code:        FAIL,
		MessageCode: messageCode,
		Params:      params,
	}
}

func FailResponseCode(messageCode MessageCode, params MessageParams) Response {
	return BuildResponse(http.StatusOK, FailDataCode(messageCode, params))
}

func FailDataError(err error) ResultStruct {
	if err == nil {
		return FailData()
	}
	var messageErr MessageError
	if errors.As(err, &messageErr) {
		return FailDataCode(messageErr.Code, messageErr.Params)
	}
	return FailData()
}

func FailResponseError(err error) Response {
	return BuildResponse(http.StatusOK, FailDataError(err))
}
