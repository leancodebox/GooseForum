package controllers

import (
	"time"

	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/docs/docProjects"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/optlogger"
	array "github.com/leancodebox/GooseForum/app/bundles/collectionopt"
)

// DocsProjectListReq 项目列表请求
type DocsProjectListReq struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Keyword  string `json:"keyword"`
	Status   *int8  `json:"status"`
	IsPublic *int8  `json:"isPublic"`
}

// DocsProjectCreateReq 创建项目请求
type DocsProjectCreateReq struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Slug        string `json:"slug" validate:"required,min=1,max=100"`
	Description string `json:"description"`
	LogoUrl     string `json:"logoUrl"`
	Status      int8   `json:"status" validate:"oneof=1 2 3"`
	IsPublic    int8   `json:"isPublic" validate:"oneof=0 1"`
	OwnerId     uint64 `json:"ownerId"`
}

// DocsProjectUpdateReq 更新项目请求
type DocsProjectUpdateReq struct {
	Id          uint64 `json:"id" validate:"required"`
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Slug        string `json:"slug" validate:"required,min=1,max=100"`
	Description string `json:"description"`
	LogoUrl     string `json:"logoUrl"`
	Status      int8   `json:"status" validate:"oneof=1 2 3"`
	IsPublic    int8   `json:"isPublic" validate:"oneof=0 1"`
	OwnerId     uint64 `json:"ownerId"`
}

// DocsProjectDeleteReq 删除项目请求
type DocsProjectDeleteReq struct {
	Id uint64 `json:"id" validate:"required"`
}

// DocsProjectItem 项目列表项
type DocsProjectItem struct {
	Id          uint64 `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	LogoUrl     string `json:"logoUrl"`
	Status      int8   `json:"status"`
	IsPublic    int8   `json:"isPublic"`
	OwnerId     uint64 `json:"ownerId"`
	OwnerName   string `json:"ownerName"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// AdminDocsProjectList 获取项目列表
func AdminDocsProjectList(req component.BetterRequest[DocsProjectListReq]) component.Response {
	params := req.Params
	
	// 设置默认值
	page := max(params.Page, 1)
	pageSize := params.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// 获取项目列表
	projects, total, err := docProjects.GetProjectList(page, pageSize, params.Keyword, params.Status, params.IsPublic)
	if err != nil {
		return component.FailResponse("获取项目列表失败: " + err.Error())
	}

	// 获取用户信息
	userIds := array.Map(projects, func(p docProjects.Entity) uint64 {
		return p.OwnerId
	})
	userMap := users.GetMapByIds(userIds)

	// 转换响应格式
	projectList := array.Map(projects, func(project docProjects.Entity) DocsProjectItem {
		ownerName := ""
		if user, exists := userMap[project.OwnerId]; exists && user != nil {
			ownerName = user.Username
		}
		return DocsProjectItem{
			Id:          project.Id,
			Name:        project.Name,
			Slug:        project.Slug,
			Description: project.Description,
			LogoUrl:     project.LogoUrl,
			Status:      project.Status,
			IsPublic:    project.IsPublic,
			OwnerId:     project.OwnerId,
			OwnerName:   ownerName,
			CreatedAt:   project.CreatedAt.Format(time.DateTime),
			UpdatedAt:   project.UpdatedAt.Format(time.DateTime),
		}
	})

	return component.SuccessPage(projectList, page, pageSize, total)
}

// AdminDocsProjectDetail 获取项目详情
func AdminDocsProjectDetail(req component.BetterRequest[component.Null]) component.Response {
	// 从URL参数中获取ID需要通过gin.Context，这里暂时使用固定值
	// TODO: 需要修改为从URL参数获取ID的方式
	project := docProjects.Get(1) // 临时使用固定ID
	if project.Id == 0 {
		return component.FailResponse("项目不存在")
	}

	// 获取用户信息
	user, _ := users.Get(project.OwnerId)
	ownerName := ""
	if user.Id != 0 {
		ownerName = user.Username
	}

	response := DocsProjectItem{
		Id:          project.Id,
		Name:        project.Name,
		Slug:        project.Slug,
		Description: project.Description,
		LogoUrl:     project.LogoUrl,
		Status:      project.Status,
		IsPublic:    project.IsPublic,
		OwnerId:     project.OwnerId,
		OwnerName:   ownerName,
		CreatedAt:   project.CreatedAt.Format(time.DateTime),
		UpdatedAt:   project.UpdatedAt.Format(time.DateTime),
	}

	return component.SuccessResponse(response)
}

// AdminDocsProjectCreate 创建项目
func AdminDocsProjectCreate(req component.BetterRequest[DocsProjectCreateReq]) component.Response {
	params := req.Params

	// 检查slug是否已存在
	if docProjects.ExistsBySlug(params.Slug) {
		return component.FailResponse("项目标识符已存在")
	}

	// 创建项目实体
	project := docProjects.Entity{
		Name:        params.Name,
		Slug:        params.Slug,
		Description: params.Description,
		LogoUrl:     params.LogoUrl,
		Status:      params.Status,
		IsPublic:    params.IsPublic,
		OwnerId:     params.OwnerId,
	}

	// 保存项目
	rowsAffected := docProjects.SaveOrCreateById(&project)
	if rowsAffected <= 0 {
		return component.FailResponse("创建项目失败")
	}

	// 记录操作日志
	optlogger.UserOpt(req.UserId, optlogger.CreateDocProject, project.Id, "创建文档项目: "+project.Name)

	// 获取用户信息
	user, _ := users.Get(project.OwnerId)
	ownerName := ""
	if user.Id != 0 {
		ownerName = user.Username
	}

	response := DocsProjectItem{
		Id:          project.Id,
		Name:        project.Name,
		Slug:        project.Slug,
		Description: project.Description,
		LogoUrl:     project.LogoUrl,
		Status:      project.Status,
		IsPublic:    project.IsPublic,
		OwnerId:     project.OwnerId,
		OwnerName:   ownerName,
		CreatedAt:   project.CreatedAt.Format(time.DateTime),
		UpdatedAt:   project.UpdatedAt.Format(time.DateTime),
	}

	return component.SuccessResponse(response)
}

// AdminDocsProjectUpdate 更新项目
func AdminDocsProjectUpdate(req component.BetterRequest[DocsProjectUpdateReq]) component.Response {
	params := req.Params

	// 获取原项目信息
	originalProject := docProjects.Get(params.Id)
	if originalProject.Id == 0 {
		return component.FailResponse("项目不存在")
	}

	// 检查slug是否被其他项目使用
	if params.Slug != originalProject.Slug && docProjects.ExistsBySlugExcludeId(params.Slug, params.Id) {
		return component.FailResponse("项目标识符已存在")
	}

	// 更新项目信息
	originalProject.Name = params.Name
	originalProject.Slug = params.Slug
	originalProject.Description = params.Description
	originalProject.LogoUrl = params.LogoUrl
	originalProject.Status = params.Status
	originalProject.IsPublic = params.IsPublic
	originalProject.OwnerId = params.OwnerId

	// 保存更新
	rowsAffected := docProjects.SaveOrCreateById(&originalProject)
	if rowsAffected <= 0 {
		return component.FailResponse("更新项目失败")
	}

	// 记录操作日志
	optlogger.UserOpt(req.UserId, optlogger.UpdateDocProject, originalProject.Id, "更新文档项目: "+originalProject.Name)

	// 获取用户信息
	user, _ := users.Get(originalProject.OwnerId)
	ownerName := ""
	if user.Id != 0 {
		ownerName = user.Username
	}

	response := DocsProjectItem{
		Id:          originalProject.Id,
		Name:        originalProject.Name,
		Slug:        originalProject.Slug,
		Description: originalProject.Description,
		LogoUrl:     originalProject.LogoUrl,
		Status:      originalProject.Status,
		IsPublic:    originalProject.IsPublic,
		OwnerId:     originalProject.OwnerId,
		OwnerName:   ownerName,
		CreatedAt:   originalProject.CreatedAt.Format(time.DateTime),
		UpdatedAt:   originalProject.UpdatedAt.Format(time.DateTime),
	}

	return component.SuccessResponse(response)
}

// AdminDocsProjectDelete 删除项目（软删除）
func AdminDocsProjectDelete(req component.BetterRequest[DocsProjectDeleteReq]) component.Response {
	params := req.Params

	// 获取项目信息
	project := docProjects.Get(params.Id)
	if project.Id == 0 {
		return component.FailResponse("项目不存在")
	}

	// 执行软删除
	rowsAffected := docProjects.SoftDelete(params.Id)
	if rowsAffected <= 0 {
		return component.FailResponse("删除项目失败")
	}

	// 记录操作日志
	optlogger.UserOpt(req.UserId, optlogger.DeleteDocProject, project.Id, "删除文档项目: "+project.Name)

	return component.SuccessResponse(nil)
}