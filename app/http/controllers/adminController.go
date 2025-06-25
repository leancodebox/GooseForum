package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	array "github.com/leancodebox/GooseForum/app/bundles/collectionopt"
	"github.com/leancodebox/GooseForum/app/bundles/jsonopt"
	"github.com/leancodebox/GooseForum/app/models/forum/applySheet"
	"github.com/leancodebox/GooseForum/app/models/forum/pageConfig"
	"net/http"
	"time"

	"github.com/leancodebox/GooseForum/app/models/forum/articleCategory"

	"github.com/leancodebox/GooseForum/app/models/forum/optRecord"

	"strings"

	"github.com/leancodebox/GooseForum/app/datastruct"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/forum/articles"
	"github.com/leancodebox/GooseForum/app/models/forum/role"
	"github.com/leancodebox/GooseForum/app/models/forum/rolePermissionRs"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/optlogger"
	"github.com/leancodebox/GooseForum/app/service/permission"
	"golang.org/x/exp/slices"
)

type UserListReq struct {
	Username string `json:"username"`
	UserId   uint64 `json:"userId"`
	Email    string `json:"email"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

type UserItem struct {
	UserId     uint64                              `json:"userId"`
	Username   string                              `json:"username"`
	AvatarUrl  string                              `json:"avatarUrl"`
	Email      string                              `json:"email"`
	Status     int8                                `json:"status"`
	Validate   int8                                `json:"validate"`
	Prestige   int64                               `json:"prestige"`
	RoleList   []datastruct.Option[string, uint64] `json:"roleList"`
	RoleId     uint64                              `json:"roleId,omitempty"`
	CreateTime string                              `json:"createTime"`
}

func UserList(req component.BetterRequest[UserListReq]) component.Response {
	var pageData = users.Page(users.PageQuery{
		Page:     req.Params.Page,
		PageSize: req.Params.PageSize,
		Username: req.Params.Username,
		UserId:   req.Params.UserId,
		Email:    req.Params.Email,
	})

	roleEntityList := role.AllEffective()
	roleMap := array.Slice2Map(roleEntityList, func(v *role.Entity) uint64 {
		return v.Id
	})
	list := array.Map(pageData.Data, func(t users.Entity) UserItem {
		roleEntity := roleMap[t.RoleId]
		var roleList []datastruct.Option[string, uint64]
		if roleEntity != nil {
			roleList = append(roleList, datastruct.Option[string, uint64]{
				Name:  roleEntity.RoleName,
				Value: roleEntity.Id,
			})
		}

		return UserItem{
			UserId:     t.Id,
			AvatarUrl:  t.GetWebAvatarUrl(),
			Username:   t.Username,
			Email:      t.Email,
			Status:     t.Status,
			Validate:   t.Validate,
			Prestige:   t.Prestige,
			RoleList:   roleList,
			RoleId:     t.RoleId,
			CreateTime: t.CreatedAt.Format(time.DateTime),
		}
	})
	return component.SuccessPage(
		list,
		pageData.Page,
		pageData.PageSize,
		pageData.Total,
	)
}

type EditUserReq struct {
	UserId   uint64 `json:"userId"`
	Status   int8   `json:"status"`
	Validate int8   `json:"validate"`
	RoleId   uint64 `json:"roleId"`
}

func EditUser(req component.BetterRequest[EditUserReq]) component.Response {
	params := req.Params
	user, err := users.Get(params.UserId)
	if err != nil || user.Id == 0 {
		return component.FailResponse("目标用户查询失败")
	}
	opt := false
	msg := "用户编辑"
	if user.Status != params.Status {
		msg = msg + fmt.Sprintf("[用户状态调整:%v->%v]", user.Status, params.Status)
		user.Status = params.Status
		opt = true
	}
	if user.Validate != params.Validate {
		msg = msg + fmt.Sprintf("[用户验证状态:%v->%v]", user.Status, params.Status)
		user.Validate = params.Validate
		opt = true
	}
	if user.RoleId != params.RoleId {
		msg = msg + fmt.Sprintf("[用户角色调整:%v->%v]", user.RoleId, params.RoleId)
		user.RoleId = params.RoleId
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
	Id            uint64 `json:"id"`
	Title         string `json:"title"`
	Type          int8   `json:"type"`   // 文章类型：0 博文，1教程，2问答，3分享
	UserId        uint64 `json:"userId"` //
	Username      string `json:"username"`
	ArticleStatus int8   `json:"articleStatus"` // 文章状态：0 草稿 1 发布
	ProcessStatus int8   `json:"processStatus"` // 管理状态：0 正常 1 封禁
	CreatedAt     string `json:"createdAt"`     // 改为 string 类型
	UpdatedAt     string `json:"updatedAt"`     // 改为 string 类型
}

func ArticlesList(req component.BetterRequest[ArticlesListReq]) component.Response {
	param := req.Params
	pageData := articles.Page[articles.SmallEntity](articles.PageQuery{Page: max(param.Page, 1), PageSize: param.PageSize, UserId: param.UserId})
	userIds := array.Map(pageData.Data, func(t articles.SmallEntity) uint64 {
		return t.UserId
	})
	userMap := users.GetMapByIds(userIds)
	return component.SuccessPage(
		array.Map(pageData.Data, func(t articles.SmallEntity) ArticlesInfoDto {
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
				CreatedAt:     t.CreatedAt.Format(time.DateTime),
				UpdatedAt:     t.UpdatedAt.Format(time.DateTime),
			}
		}),
		pageData.Page,
		pageData.PageSize,
		pageData.Total,
	)
}

type EditArticleReq struct {
	Id            uint64 `json:"id" validate:"required"`
	ProcessStatus int8   `json:"processStatus" validate:"oneof=0 1"` // 0正常 1封禁
}

// EditArticle 文章状态管理
func EditArticle(req component.BetterRequest[EditArticleReq]) component.Response {
	article := articles.Get(req.Params.Id)
	if article.Id == 0 {
		return component.FailResponse("文章不存在")
	}

	// 更新文章状态
	article.ProcessStatus = req.Params.ProcessStatus
	err := articles.Save(&article)
	if err != nil {
		return component.FailResponse("操作失败")
	}

	// 记录操作日志
	status := "解除封禁"
	if req.Params.ProcessStatus == 1 {
		status = "封禁"
	}
	optlogger.UserOpt(req.UserId, optlogger.EditArticle, article.Id,
		fmt.Sprintf("文章%s操作:[%s]", status, article.Title))

	return component.SuccessResponse("操作成功")
}

type PermissionListReq struct {
}

func GetPermissionList(req component.BetterRequest[PermissionListReq]) component.Response {
	res := permission.BuildOptions()
	return component.SuccessResponse(res)
}

type GetAllRoleItemReq struct {
}

func GetAllRoleItem(req component.BetterRequest[GetAllRoleItemReq]) component.Response {
	res := array.Map(role.AllEffective(), func(t *role.Entity) datastruct.Option[string, uint64] {
		return datastruct.Option[string, uint64]{Name: t.RoleName, Label: t.RoleName, Value: t.Id}
	})
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
	roleIds := array.Map(pageData.Data, func(t role.Entity) uint64 {
		return t.Id
	})
	rpGroup := make(map[uint64][]uint64)
	if len(roleIds) > 0 {
		rpGroup = rolePermissionRs.GetRsGroupByRoleIds(roleIds)
	}
	list := array.Map(pageData.Data, func(t role.Entity) RoleItem {
		pList, ok := rpGroup[t.Id]
		permissionItemList := make([]PermissionItem, 0)
		if ok {
			permissionItemList = array.Map(pList, func(t uint64) PermissionItem {
				p := permission.Enum(t)
				return PermissionItem{Id: p.Id(), Name: p.Name()}
			})
		}
		return RoleItem{
			RoleId:      t.Id,
			RoleName:    t.RoleName,
			Effective:   t.Effective,
			Permissions: permissionItemList,
			CreateTime:  t.CreatedAt.Format(time.DateTime),
		}
	})

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
			item.Effective = 1
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

type OptRecordPageReq struct {
}

func OptRecordPage(req component.BetterRequest[OptRecordPageReq]) component.Response {
	pageData := optRecord.Page(optRecord.PageQuery{})
	return component.SuccessPage(
		array.Map(pageData.Data, func(item optRecord.Entity) optRecord.Entity {
			return item
		}),
		pageData.Page,
		pageData.PageSize,
		pageData.Total,
	)
}

type CategoryListReq struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

type CategoryItem struct {
	Id       uint64 `json:"id"`
	Category string `json:"category"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status"`
}

// GetCategoryList 获取分类列表
func GetCategoryList(req component.BetterRequest[CategoryListReq]) component.Response {
	categories := articleCategory.All()
	return component.SuccessResponse(array.Map(categories, func(t *articleCategory.Entity) CategoryItem {
		return CategoryItem{
			Id:       t.Id,
			Category: t.Category,
			//Sort:     t.Sort,
			//Status:   t.Status,
		}
	}))
}

type CategorySaveReq struct {
	Id       uint64 `json:"id"`
	Category string `json:"category" validate:"required"`
	Sort     int    `json:"sort"`
	Status   int8   `json:"status"`
}

// SaveCategory 保存分类
func SaveCategory(req component.BetterRequest[CategorySaveReq]) component.Response {
	if len(strings.TrimSpace(req.Params.Category)) == 0 {
		return component.FailResponse("分类名称不能为空")
	}

	entity := articleCategory.Get(req.Params.Id)
	if req.Params.Id != 0 && entity.Id == 0 {
		return component.FailResponse("数据不存在")
	}
	entity.Category = req.Params.Category

	articleCategory.SaveOrCreateById(&entity)
	return component.SuccessResponse(true)
}

// DeleteCategory 删除分类
func DeleteCategory(req component.BetterRequest[struct {
	Id uint64 `json:"id"`
}]) component.Response {
	entity := articleCategory.Get(req.Params.Id)
	if entity.Id == 0 {
		return component.FailResponse("分类不存在")
	}
	if articleCategory.Count() == 1 {
		return component.FailResponse("至少保留1个分类")
	}
	articleCategory.DeleteEntity(&entity)
	return component.SuccessResponse(true)
}

type ApplySheetListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Search   string `form:"search"`
	UserId   uint64 `form:"userId"`
}

func ApplySheet(req component.BetterRequest[ApplySheetListReq]) component.Response {
	pageData := applySheet.Page[applySheet.Entity](applySheet.PageQuery{
		Page:     req.Params.Page,
		PageSize: req.Params.PageSize,
	})

	return component.SuccessPage(array.Map(pageData.Data, func(item applySheet.Entity) applySheet.Entity {
		return item
	}),
		pageData.Page,
		pageData.PageSize,
		pageData.Total,
	)
}

const (
	FriendShipLinks = `friendShipLinks`
)

var pageTypeList = []string{
	FriendShipLinks,
}

func GetPageConfig(c *gin.Context) {
	pageType := c.Param(`pageType`)
	if !slices.Contains(pageTypeList, pageType) {
		c.JSON(http.StatusOK, component.FailData(`类型不存在`))
		return
	}
	configEntity := pageConfig.GetByPageType(pageType)

	c.JSON(http.StatusOK, component.SuccessData(map[string]any{
		`pageType`: pageType,
		`config`:   configEntity.Config,
	}))
}

type FriendLinksGroup struct {
	Name  string     `json:"name,omitempty"`
	Links []LinkItem `json:"links,omitempty"`
}

func GetFriendLinks(req component.BetterRequest[null]) component.Response {
	configEntity := pageConfig.GetByPageType(FriendShipLinks)
	var res []FriendLinksGroup
	if configEntity.Id == 0 {
		lItem := LinkItem{
			Name:    "GooseForum",
			Desc:    "简单的社区构建软件 / Easy forum software for building friendly communities.",
			Url:     "https://gooseforum.online",
			LogoUrl: "/static/pic/default-avatar.png",
		}
		res = []FriendLinksGroup{
			FriendLinksGroup{
				Name:  "community",
				Links: []LinkItem{lItem},
			},
			FriendLinksGroup{
				Name:  "blog",
				Links: []LinkItem{lItem},
			},
			FriendLinksGroup{
				Name:  "tool",
				Links: []LinkItem{lItem},
			},
		}
	} else {
		res = jsonopt.Decode[[]FriendLinksGroup](configEntity.Config)
	}
	return component.SuccessResponse(res)
}

type SaveFriendLinksReq struct {
	LinksInfo []FriendLinksGroup `json:"linksInfo"`
}

func SaveFriendLinks(req component.BetterRequest[SaveFriendLinksReq]) component.Response {
	configEntity := pageConfig.GetByPageType(FriendShipLinks)
	configEntity.PageType = FriendShipLinks
	configEntity.Config = jsonopt.Encode(req.Params.LinksInfo)
	pageConfig.CreateOrSave(&configEntity)
	return component.SuccessResponse("success")
}
