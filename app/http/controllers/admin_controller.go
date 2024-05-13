package controllers

import (
	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/userRoleRs"
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
	UserId     uint64                              `json:"userId"`
	Username   string                              `json:"username"`
	Email      string                              `json:"email"`
	RoleList   []datastruct.Option[string, uint64] `json:"roleList"`
	CreateTime string                              `json:"createTime"`
}

func UserList(req component.BetterRequest[UserListReq]) component.Response {
	var pageData = users.Page(users.PageQuery{
		Username: req.Params.Username,
		UserId:   req.Params.UserId,
		Email:    req.Params.Email,
	})

	userIdList := array.ArrayMap(func(t users.Entity) uint64 {
		return t.Id
	}, pageData.Data)
	userRoleMap := userRoleRs.GetRoleGroupByUserIds(userIdList)
	list := array.ArrayMap(func(t users.Entity) UserItem {
		var roleList []datastruct.Option[string, uint64]
		if data, ok := userRoleMap[t.Id]; ok {
			roleList = array.ArrayMap(func(rItem role.Entity) datastruct.Option[string, uint64] {
				return datastruct.Option[string, uint64]{
					Name:  rItem.RoleName,
					Value: rItem.Id,
				}
			}, data)
		}
		return UserItem{
			UserId:     t.Id,
			Username:   t.Username,
			Email:      t.Email,
			RoleList:   roleList,
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
	return component.SuccessResponse("")
}

type ArticlesListReq struct {
}

func ArticlesList(req component.BetterRequest[ArticlesListReq]) component.Response {
	return component.SuccessResponse("")
}

type EditArticleReq struct {
}

func EditArticle(req component.BetterRequest[EditArticleReq]) component.Response {
	return component.SuccessResponse("")
}

type RoleListReq struct {
}

type RoleItem struct {
	RoleId      uint64           `json:"roleId"`
	RoleName    string           `json:"roleName"`
	Permissions []PermissionItem `json:"permissions"`
	CreateTime  string           `json:"createTime"`
}

type PermissionItem struct {
	Id   uint64 `json:"id"`
	Name string `json:"name"`
}

func RoleList(req component.BetterRequest[RoleListReq]) component.Response {
	pageData := role.Page(role.PageQuery{})
	roleIds := array.ArrayMap(func(t role.Entity) uint64 {
		return t.Id
	}, pageData.Data)
	rpGroup := make(map[uint64][]uint64)
	if len(roleIds) > 0 {
		rpGroup = rolePermissionRs.GetRsGroupByRoleIds(roleIds)
	}
	list := array.ArrayMap(func(t role.Entity) RoleItem {
		pList, ok := rpGroup[t.Id]
		permissionItemList := make([]PermissionItem, 0)
		if ok {
			permissionItemList = array.ArrayMap(func(t uint64) PermissionItem {
				p := permission.Enum(t)
				return PermissionItem{Id: p.Id(), Name: p.Name()}
			}, pList)
		}
		return RoleItem{
			RoleId:      t.Id,
			RoleName:    t.RoleName,
			Permissions: permissionItemList,
			CreateTime:  t.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}, pageData.Data)

	return component.SuccessResponse(component.DataMap{
		"list":  list,
		"size":  pageData.PageSize,
		"total": pageData.Total,
	})
}

type RoleSaveReq struct {
}

func RoleSave(req component.BetterRequest[RoleSaveReq]) component.Response {
	return component.SuccessResponse("")
}
