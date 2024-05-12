package controllers

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/permission"
	array "github.com/leancodebox/goose/collectionopt"
)

type UserListReq struct {
	Username string `json:"username"`
	UserId   uint64 `json:"userId"`
	Email    string `json:"email"`
}

type UserItem struct {
	UserId     uint64 `json:"userId"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	CreateTime string `json:"createTime"`
}

func UserList(req component.BetterRequest[UserListReq]) component.Response {
	user, err := req.GetUser()
	if err != nil {
		return component.FailResponse(err.Error())
	}
	if permission.CheckUser(user.Id, permission.UserManager) == false {
		return component.FailResponse("权限不足")
	}
	var pageData = users.Page(users.PageQuery{
		Username: req.Params.Username,
		UserId:   req.Params.UserId,
		Email:    req.Params.Email,
	})
	list := array.ArrayMap(func(t users.Entity) UserItem {
		return UserItem{
			UserId:     t.Id,
			Username:   t.Username,
			Email:      t.Email,
			CreateTime: t.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}, pageData.Data)
	return component.SuccessResponse(map[string]any{
		"list":  list,
		"size":  pageData.PageSize,
		"total": pageData.Total,
	})

}

type EditUserReq struct {
}

func EditUser(req component.BetterRequest[EditUserReq]) component.Response {
	user, err := req.GetUser()
	if err != nil {
		return component.FailResponse(err.Error())
	}
	if permission.CheckUser(user.Id, permission.UserManager) == false {
		return component.FailResponse("权限不足")
	}
	return component.SuccessResponse("")

}

type ArticlesListReq struct {
}

func ArticlesList(req component.BetterRequest[ArticlesListReq]) component.Response {
	user, err := req.GetUser()
	if err != nil {
		return component.FailResponse(err.Error())
	}
	if permission.CheckUser(user.Id, permission.ArticlesManager) == false {
		return component.FailResponse("权限不足")
	}
	return component.SuccessResponse("")

}

type EditArticleReq struct {
}

func EditArticle(req component.BetterRequest[EditArticleReq]) component.Response {
	user, err := req.GetUser()
	if err != nil {
		return component.FailResponse(err.Error())
	}
	if permission.CheckUser(user.Id, permission.ArticlesManager) == false {
		return component.FailResponse("权限不足")
	}
	return component.SuccessResponse("")

}
