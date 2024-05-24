package controllers

import (
	"fmt"
	"time"

	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/userRoleRs"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/optlogger"
	"github.com/leancodebox/GooseForum/app/service/permission"
	array "github.com/leancodebox/goose/collectionopt"
	"golang.org/x/exp/slices"
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
	Status     int8                                `json:"status"`
	Validate   int8                                `json:"validate"`
	Prestige   int64                               `json:"prestige"`
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
			Status:     t.Status,
			Validate:   t.Validate,
			Prestige:   t.Prestige,
			RoleList:   roleList,
			CreateTime: t.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}, pageData.Data)
	return component.SuccessPage(
		list,
		pageData.Page,
		pageData.PageSize,
		pageData.Total,
	)
}

type EditUserReq struct {
	UserId uint64 `json:"userId"`
	Status int8   `json:"status"`
}

func EditUser(req component.BetterRequest[EditUserReq]) component.Response {
	params := req.Params
	user, err := users.Get(params.UserId)
	if err != nil || user.Id == 0 {
		return component.SuccessResponse("目标用户查询失败")
	}
	opt := false
	msg := "用户编辑"
	if user.Status != params.Status {
		msg = msg + fmt.Sprintf("[用户状态调整:%v->%v]", user.Status, params.Status)
		user.Status = params.Status
		opt = true
	}
	if opt {
		users.Save(&user)
		optlogger.UserOpt(req.UserId, optlogger.EditUser, user.Id, msg)
	}
	return component.SuccessResponse("success")
}

type ArticlesListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Search   string `form:"search"`
	UserId   uint64 `form:"userId"`
}

type ArticlesInfoDto struct {
	Id            uint64    `json:"id"`
	Title         string    `json:"title"`
	Type          int8      ` json:"type"`  // 文章类型：0 博文，1教程，2问答，3分享
	UserId        uint64    `json:"userId"` //
	Username      string    `json:"username"`
	ArticleStatus int8      `json:"articleStatus"` // 文章状态：0 草稿 1 发布
	ProcessStatus int8      `json:"processStatus"` // 管理状态：0 正常 1 封禁
	CreatedAt     time.Time ` json:"createdAt"`    //
	UpdatedAt     time.Time ` json:"updatedAt"`    //
}

func ArticlesList(req component.BetterRequest[ArticlesListReq]) component.Response {
	param := req.Params
	pageData := articles.Page(articles.PageQuery{Page: max(param.Page, 1), PageSize: param.PageSize, UserId: param.UserId})
	userIds := array.ArrayMap(func(t articles.Entity) uint64 {
		return t.UserId
	}, pageData.Data)
	userMap := users.GetMapByIds(userIds)
	return component.SuccessPage(
		array.ArrayMap(func(t articles.Entity) ArticlesInfoDto {
			username := ""
			if user, _ := userMap[t.UserId]; user != nil {
				username = user.Username
			}
			return ArticlesInfoDto{
				Id:            t.Id,
				Title:         t.Title,
				Type:          t.Type,
				UserId:        t.UserId,
				Username:      username,
				ArticleStatus: t.ArticleStatus,
				ProcessStatus: t.ProcessStatus,
				CreatedAt:     t.CreatedAt,
				UpdatedAt:     t.UpdatedAt,
			}
		}, pageData.Data),
		pageData.Page,
		pageData.PageSize,
		pageData.Total,
	)
}

type EditArticleReq struct {
}

// EditArticle 冻结操作
func EditArticle(req component.BetterRequest[EditArticleReq]) component.Response {
	return component.SuccessResponse("")
}

type PermissionListReq struct {
}

func GetPermissionList(req component.BetterRequest[PermissionListReq]) component.Response {
	res := permission.BuildOptions()
	return component.SuccessResponse(res)
}

type RoleListReq struct {
}

type RoleItem struct {
	RoleId      uint64           `json:"roleId"`
	RoleName    string           `json:"roleName"`
	Effective   int              `json:"effective"`
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
			Effective:   t.Effective,
			Permissions: permissionItemList,
			CreateTime:  t.CreatedAt.Format("2006-01-02 15:04:05"),
		}
	}, pageData.Data)

	return component.SuccessPage(
		list,
		pageData.Page,
		pageData.PageSize,
		pageData.Total,
	)
}

type RoleSaveReq struct {
	Id          uint     `json:"id"`
	RoleName    string   `json:"roleName" validate:"required"`
	Permissions []uint64 `json:"permissions" validate:"required,min=1,max=100"`
}

func RoleSave(req component.BetterRequest[RoleSaveReq]) component.Response {
	var roleEntity role.Entity
	if req.Params.Id > 0 {
		roleEntity = role.Get(req.Params.Id)
	} else {
		roleEntity = role.Entity{
			Effective: 1,
		}
	}
	roleEntity.RoleName = req.Params.RoleName
	role.SaveOrCreateById(&roleEntity)

	rsList := rolePermissionRs.GetRsByRoleId(roleEntity.Id)
	var canUpdate []uint64
	// 更新数据
	for _, item := range rsList {
		item.Effective = 0
		if slices.Contains(req.Params.Permissions, item.PermissionId) {
			item.Effective = 0
			canUpdate = append(canUpdate, item.PermissionId)
		}
		rolePermissionRs.SaveOrCreateById(item)
	}
	// 删除数据
	for _, item := range req.Params.Permissions {
		if !slices.Contains(canUpdate, item) {
			rsItem := rolePermissionRs.Entity{
				RoleId:       roleEntity.Id,
				PermissionId: item,
				Effective:    1,
			}
			rolePermissionRs.SaveOrCreateById(&rsItem)
		}
	}

	return component.SuccessResponse(true)
}

type RoleSaveDel struct {
	Id uint `json:"id"`
}

func RoleDel(req component.BetterRequest[RoleSaveDel]) component.Response {
	roleEntity := role.Get(req.Params.Id)
	if roleEntity.Id == 0 {
		return component.FailResponse("角色不存在")
	}
	rsList := rolePermissionRs.GetRsByRoleId(roleEntity.Id)
	// 删除
	for _, item := range rsList {
		rolePermissionRs.DeleteEntity(item)
	}
	role.DeleteEntity(&roleEntity)
	return component.SuccessResponse(true)
}
