package api

import (
	"errors"
	"net/http"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/service/postrevisionservice"
)

type PostRevisionListReq struct {
	PostID uint64 `form:"postId" validate:"required"`
	Cursor uint64 `form:"cursor"`
	Limit  uint64 `form:"limit"`
}

type PostRevisionDetailReq struct {
	PostID     uint64 `form:"postId" validate:"required"`
	RevisionNo uint64 `form:"revisionNo" validate:"required"`
}

func ListPostRevisions(req component.BetterRequest[PostRevisionListReq]) component.Response {
	page, err := postrevisionservice.List(req.UserId, req.Params.PostID, req.Params.Cursor, req.Params.Limit)
	if err != nil {
		return postRevisionError(err)
	}
	return component.SuccessResponse(page)
}

func GetPostRevision(req component.BetterRequest[PostRevisionDetailReq]) component.Response {
	revision, err := postrevisionservice.Get(req.UserId, req.Params.PostID, req.Params.RevisionNo)
	if err != nil {
		return postRevisionError(err)
	}
	return component.SuccessResponse(revision)
}

func postRevisionError(err error) component.Response {
	if errors.Is(err, postrevisionservice.ErrNotFound) {
		return component.BuildResponse(http.StatusNotFound, component.FailDataCode(component.MessagePostNotFound, nil))
	}
	if errors.Is(err, postrevisionservice.ErrForbidden) {
		return component.BuildResponse(http.StatusForbidden, component.FailDataCode(component.MessagePermissionDenied, nil))
	}
	return component.FailResponseCode(component.MessageOperationFailed, nil)
}
