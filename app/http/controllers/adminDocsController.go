package controllers

import (
	"fmt"
	"time"

	"github.com/spf13/cast"

	array "github.com/leancodebox/GooseForum/app/bundles/collectionopt"
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/app/models/docs/docContents"
	"github.com/leancodebox/GooseForum/app/models/docs/docProjects"
	"github.com/leancodebox/GooseForum/app/models/docs/docVersions"
	"github.com/leancodebox/GooseForum/app/models/forum/users"
	"github.com/leancodebox/GooseForum/app/service/optlogger"
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

// AdminDocsContentCreate 创建内容
func AdminDocsContentCreate(req component.BetterRequest[DocsContentCreateReq]) component.Response {
	params := req.Params

	// 检查版本是否存在
	version := docVersions.Get(params.VersionId)
	if version.Id == 0 {
		return component.FailResponse("版本不存在")
	}

	// 检查slug在版本内是否已存在
	if docContents.ExistsBySlugAndVersionId(params.Slug, params.VersionId) {
		return component.FailResponse("该版本下内容标识符已存在")
	}

	// 创建内容
	content := &docContents.Entity{
		VersionId:   params.VersionId,
		Title:       params.Title,
		Slug:        params.Slug,
		Content:     params.Content,
		SortOrder:   params.SortOrder,
		IsPublished: 0, // 默认为草稿
	}

	rowsAffected := docContents.SaveOrCreateById(content)
	if rowsAffected == 0 {
		return component.FailResponse("创建内容失败")
	}

	// 记录操作日志
	optlogger.UserOpt(req.UserId, optlogger.CreateDocContent, content.Id, fmt.Sprintf("创建文档内容: %s", params.Title))

	return component.SuccessResponse(map[string]interface{}{
		"id": content.Id,
	})
}

// AdminDocsContentUpdate 更新内容
func AdminDocsContentUpdate(req component.BetterRequest[DocsContentUpdateReq]) component.Response {
	params := req.Params

	// 获取原内容信息
	originalContent := docContents.Get(params.Id)
	if originalContent.Id == 0 {
		return component.FailResponse("内容不存在")
	}

	// 检查版本是否存在
	version := docVersions.Get(params.VersionId)
	if version.Id == 0 {
		return component.FailResponse("版本不存在")
	}

	// 检查slug在版本内是否被其他内容使用
	if params.Slug != originalContent.Slug && docContents.ExistsBySlugAndVersionIdExcludeId(params.Slug, params.VersionId, params.Id) {
		return component.FailResponse("该版本下内容标识符已被其他内容使用")
	}

	// 更新内容
	updatedContent := &docContents.Entity{
		Id:        params.Id,
		VersionId: params.VersionId,
		Title:     params.Title,
		Slug:      params.Slug,
		Content:   params.Content,
		SortOrder: params.SortOrder,
		// 保持原有的发布状态
		IsPublished: originalContent.IsPublished,
	}
	fmt.Println(updatedContent)
	rowsAffected := docContents.SaveOrCreateById(updatedContent)
	if rowsAffected == 0 {
		return component.FailResponse("更新内容失败")
	}

	// 记录操作日志
	optlogger.UserOpt(req.UserId, optlogger.UpdateDocContent, params.Id, fmt.Sprintf("更新文档内容: %s", params.Title))

	return component.SuccessResponse("内容更新成功")
}

// AdminDocsContentDelete 删除内容
func AdminDocsContentDelete(req component.BetterRequest[DocsContentDeleteReq]) component.Response {
	params := req.Params

	// 获取内容信息
	content := docContents.Get(params.Id)
	if content.Id == 0 {
		return component.FailResponse("内容不存在")
	}

	// 删除内容
	rowsAffected := docContents.Delete(params.Id)
	if rowsAffected == 0 {
		return component.FailResponse("删除内容失败")
	}

	// 记录操作日志
	optlogger.UserOpt(req.UserId, optlogger.DeleteDocContent, params.Id, "删除文档内容")

	return component.SuccessResponse("内容删除成功")
}

// AdminDocsContentPublish 发布内容
func AdminDocsContentPublish(req component.BetterRequest[DocsContentPublishReq]) component.Response {
	params := req.Params

	// 获取内容信息
	content := docContents.Get(params.Id)
	if content.Id == 0 {
		return component.FailResponse("内容不存在")
	}

	// 更新发布状态
	rowsAffected := docContents.UpdatePublishStatus(params.Id, 1)
	if rowsAffected == 0 {
		return component.FailResponse("发布内容失败")
	}

	// 记录操作日志
	optlogger.UserOpt(req.UserId, optlogger.PublishDocContent, params.Id, "发布文档内容")

	return component.SuccessResponse("内容发布成功")
}

// AdminDocsContentDraft 设为草稿
func AdminDocsContentDraft(req component.BetterRequest[DocsContentDraftReq]) component.Response {
	params := req.Params

	// 获取内容信息
	content := docContents.Get(params.Id)
	if content.Id == 0 {
		return component.FailResponse("内容不存在")
	}

	// 更新发布状态
	rowsAffected := docContents.UpdatePublishStatus(params.Id, 0)
	if rowsAffected == 0 {
		return component.FailResponse("设为草稿失败")
	}

	// 记录操作日志
	optlogger.UserOpt(req.UserId, optlogger.DraftDocContent, params.Id, "设为草稿")

	return component.SuccessResponse("内容已设为草稿")
}

// AdminDocsContentPreview 预览内容
func AdminDocsContentPreview(req component.BetterRequest[DocsContentPreviewReq]) component.Response {
	params := req.Params

	// TODO: 这里需要集成Markdown渲染器
	// 暂时返回原始内容和空的目录
	response := DocsContentPreviewResponse{
		Html: "<p>" + params.Content + "</p>", // 临时处理
		Toc:  "",                              // 临时为空
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

// ==================== 内容管理模块 ====================

// DocsContentListReq 内容列表请求
type DocsContentListReq struct {
	Page      int     `json:"page"`
	PageSize  int     `json:"pageSize"`
	VersionId *uint64 `json:"versionId"`
	Keyword   string  `json:"keyword"`
	Status    *int8   `json:"status"` // 0:草稿 1:已发布
}

// DocsContentCreateReq 创建内容请求
type DocsContentCreateReq struct {
	VersionId uint64 `json:"versionId" validate:"required"`
	Title     string `json:"title" validate:"required,min=1,max=200"`
	Slug      string `json:"slug" validate:"required,min=1,max=200"`
	Content   string `json:"content"`
	SortOrder int    `json:"sortOrder"`
}

// DocsContentUpdateReq 更新内容请求
type DocsContentUpdateReq struct {
	Id        uint64 `json:"id" validate:"required"`
	VersionId uint64 `json:"versionId" validate:"required"`
	Title     string `json:"title" validate:"required,min=1,max=200"`
	Slug      string `json:"slug" validate:"required,min=1,max=200"`
	Content   string `json:"content"`
	SortOrder int    `json:"sortOrder"`
}

// DocsContentDeleteReq 删除内容请求
type DocsContentDeleteReq struct {
	Id uint64 `json:"id" validate:"required"`
}

// DocsContentPublishReq 发布内容请求
type DocsContentPublishReq struct {
	Id uint64 `json:"id" validate:"required"`
}

// DocsContentDraftReq 设为草稿请求
type DocsContentDraftReq struct {
	Id uint64 `json:"id" validate:"required"`
}

// DocsContentPreviewReq 预览内容请求
type DocsContentPreviewReq struct {
	Content string `json:"content" validate:"required"`
}

// DocsContentItem 内容项响应
type DocsContentItem struct {
	Id          uint64 `json:"id"`
	VersionId   uint64 `json:"versionId"`
	VersionName string `json:"versionName"`
	ProjectId   uint64 `json:"projectId"`
	ProjectName string `json:"projectName"`
	Title       string `json:"title"`
	Slug        string `json:"slug"`
	Content     string `json:"content"`
	IsPublished int8   `json:"isPublished"`
	SortOrder   int    `json:"sortOrder"`
	ViewCount   uint64 `json:"viewCount"`
	LikeCount   uint64 `json:"likeCount"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// DocsContentPreviewResponse 预览响应
type DocsContentPreviewResponse struct {
	Html string `json:"html"`
	Toc  string `json:"toc"`
}

// AdminDocsContentList 内容列表
func AdminDocsContentList(req component.BetterRequest[DocsContentListReq]) component.Response {
	params := req.Params

	// 设置默认值
	page := max(params.Page, 1)
	pageSize := params.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// 获取内容列表
	versionId := uint64(0)
	if params.VersionId != nil {
		versionId = *params.VersionId
	}
	status := -1
	if params.Status != nil {
		status = int(*params.Status)
	}
	contents, total, err := docContents.GetContentList(page, pageSize, versionId, params.Keyword, status)
	if err != nil {
		return component.FailResponse("获取内容列表失败: " + err.Error())
	}

	// 转换为响应格式
	var items []DocsContentItem
	for _, content := range contents {
		// 获取版本信息
		version := docVersions.Get(content.VersionId)
		versionName := ""
		projectId := uint64(0)
		projectName := ""
		if version.Id != 0 {
			versionName = version.Name
			projectId = version.ProjectId
			// 获取项目信息
			project := docProjects.Get(version.ProjectId)
			if project.Id != 0 {
				projectName = project.Name
			}
		}

		item := DocsContentItem{
			Id:          content.Id,
			VersionId:   content.VersionId,
			VersionName: versionName,
			ProjectId:   projectId,
			ProjectName: projectName,
			Title:       content.Title,
			Slug:        content.Slug,
			Content:     content.Content,
			IsPublished: content.IsPublished,
			SortOrder:   content.SortOrder,
			ViewCount:   content.ViewCount,
			LikeCount:   content.LikeCount,
			CreatedAt:   content.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:   content.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		items = append(items, item)
	}

	response := map[string]interface{}{
		"list":     items,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}

	return component.SuccessResponse(response)
}

// AdminDocsContentDetail 内容详情
func AdminDocsContentDetail(req component.BetterRequest[component.Null]) component.Response {
	// 从URL参数中获取ID
	id := req.GinContext.Param("id")
	if id == "" {
		return component.FailResponse("内容ID不能为空")
	}

	content := docContents.GetByIdString(id)
	if content.Id == 0 {
		return component.FailResponse("内容不存在")
	}

	// 获取版本信息
	version := docVersions.Get(content.VersionId)
	versionName := ""
	projectId := uint64(0)
	projectName := ""
	if version.Id != 0 {
		versionName = version.Name
		projectId = version.ProjectId
		// 获取项目信息
		project := docProjects.Get(version.ProjectId)
		if project.Id != 0 {
			projectName = project.Name
		}
	}

	response := DocsContentItem{
		Id:          content.Id,
		VersionId:   content.VersionId,
		VersionName: versionName,
		ProjectId:   projectId,
		ProjectName: projectName,
		Title:       content.Title,
		Slug:        content.Slug,
		Content:     content.Content,
		IsPublished: content.IsPublished,
		SortOrder:   content.SortOrder,
		ViewCount:   content.ViewCount,
		LikeCount:   content.LikeCount,
		CreatedAt:   content.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   content.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	return component.SuccessResponse(response)
}

// ============ 版本管理相关 ============

// DocsVersionListReq 版本列表请求
type DocsVersionListReq struct {
	Page      int     `json:"page"`
	PageSize  int     `json:"pageSize"`
	ProjectId *uint64 `json:"projectId"`
	Keyword   string  `json:"keyword"`
	Status    int8    `json:"status"`
}

// DocsVersionCreateReq 创建版本请求
type DocsVersionCreateReq struct {
	ProjectId   uint64 `json:"projectId" validate:"required"`
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Slug        string `json:"slug" validate:"required,min=1,max=100"`
	Description string `json:"description"`
	Status      int8   `json:"status" validate:"oneof=1 2 3"`
	IsDefault   int8   `json:"isDefault" validate:"oneof=0 1"`
	SortOrder   int    `json:"sortOrder"`
}

// DocsVersionUpdateReq 更新版本请求
type DocsVersionUpdateReq struct {
	Id          uint64 `json:"id" validate:"required"`
	ProjectId   uint64 `json:"projectId" validate:"required"`
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Slug        string `json:"slug" validate:"required,min=1,max=100"`
	Description string `json:"description"`
	Status      int8   `json:"status" validate:"oneof=1 2 3"`
	IsDefault   int8   `json:"isDefault" validate:"oneof=0 1"`
	SortOrder   int    `json:"sortOrder"`
}

// DocsVersionDeleteReq 删除版本请求
type DocsVersionDeleteReq struct {
}

// DocsVersionSetDefaultReq 设置默认版本请求
type DocsVersionSetDefaultReq struct {
	Id uint64 `json:"id" validate:"required"`
}

// DocsVersionDirectoryUpdateReq 更新目录结构请求
type DocsVersionDirectoryUpdateReq struct {
	DirectoryStructure []docVersions.DirectoryItem `json:"directoryStructure"`
}

// DocsVersionItem 版本列表项
type DocsVersionItem struct {
	Id                 uint64                      `json:"id"`
	ProjectId          uint64                      `json:"projectId"`
	ProjectName        string                      `json:"projectName"`
	Name               string                      `json:"name"`
	Slug               string                      `json:"slug"`
	Description        string                      `json:"description"`
	Status             int8                        `json:"status"`
	IsDefault          int8                        `json:"isDefault"`
	SortOrder          int                         `json:"sortOrder"`
	DirectoryStructure []docVersions.DirectoryItem `json:"directoryStructure"`
	CreatedAt          string                      `json:"createdAt"`
	UpdatedAt          string                      `json:"updatedAt"`
}

// AdminDocsVersionList 获取版本列表
func AdminDocsVersionList(req component.BetterRequest[DocsVersionListReq]) component.Response {
	params := req.Params

	// 设置默认值
	page := max(params.Page, 1)
	pageSize := params.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// 获取版本列表
	projectId := uint64(0)
	if params.ProjectId != nil {
		projectId = *params.ProjectId
	}
	status := -1
	if params.Status != 0 {
		status = int(params.Status)
	}
	versions, total, err := docVersions.GetVersionList(page, pageSize, projectId, params.Keyword, status)
	if err != nil {
		return component.FailResponse("获取版本列表失败: " + err.Error())
	}

	// 获取项目信息
	projectIds := array.Map(versions, func(v docVersions.Entity) uint64 {
		return v.ProjectId
	})
	projectMap := make(map[uint64]string)
	for _, itemProjectId := range projectIds {
		project := docProjects.Get(itemProjectId)
		if project.Id != 0 {
			projectMap[itemProjectId] = project.Name
		}
	}
	versionIds := array.Map(versions, func(version docVersions.Entity) uint64 {
		return version.Id
	})
	contentGroup := array.GroupBy(docContents.GetByVersionIds(versionIds), func(t *docContents.SimpleEntity) uint64 {
		return t.VersionId
	})

	// 转换响应格式
	versionList := array.Map(versions, func(version docVersions.Entity) DocsVersionItem {
		projectName := ""
		if name, exists := projectMap[version.ProjectId]; exists {
			projectName = name
		}

		versionContent := contentGroup[version.Id]
		contentList := array.Map(versionContent, func(t *docContents.SimpleEntity) docVersions.DirectoryItem {
			return docVersions.DirectoryItem{
				Title:       t.Title,
				Slug:        t.Slug,
				Description: "",
				Children:    nil,
			}
		})
		resDirectory := docVersions.BuildSafeDescription(version.Directory, contentList)
		return DocsVersionItem{
			Id:                 version.Id,
			ProjectId:          version.ProjectId,
			ProjectName:        projectName,
			Name:               version.Name,
			Slug:               version.Slug,
			Description:        version.Description,
			Status:             version.Status,
			IsDefault:          version.IsDefault,
			SortOrder:          version.SortOrder,
			DirectoryStructure: resDirectory,
			CreatedAt:          version.CreatedAt.Format(time.DateTime),
			UpdatedAt:          version.UpdatedAt.Format(time.DateTime),
		}
	})

	return component.SuccessPage(versionList, page, pageSize, total)
}

// AdminDocsVersionDetail 获取版本详情
func AdminDocsVersionDetail(req component.BetterRequest[component.Null]) component.Response {
	// 从URL参数中获取ID
	id := req.GinContext.Param("id")
	if id == "" {
		return component.FailResponse("版本ID不能为空")
	}

	version := docVersions.GetByIdString(id)
	if version.Id == 0 {
		return component.FailResponse("版本不存在")
	}

	// 获取项目信息
	project := docProjects.Get(version.ProjectId)
	projectName := ""
	if project.Id != 0 {
		projectName = project.Name
	}

	response := DocsVersionItem{
		Id:                 version.Id,
		ProjectId:          version.ProjectId,
		ProjectName:        projectName,
		Name:               version.Name,
		Slug:               version.Slug,
		Description:        version.Description,
		Status:             version.Status,
		IsDefault:          version.IsDefault,
		SortOrder:          version.SortOrder,
		DirectoryStructure: version.Directory,
		CreatedAt:          version.CreatedAt.Format(time.DateTime),
		UpdatedAt:          version.UpdatedAt.Format(time.DateTime),
	}

	return component.SuccessResponse(response)
}

// AdminDocsVersionCreate 创建版本
func AdminDocsVersionCreate(req component.BetterRequest[DocsVersionCreateReq]) component.Response {
	params := req.Params

	// 检查项目是否存在
	project := docProjects.Get(params.ProjectId)
	if project.Id == 0 {
		return component.FailResponse("项目不存在")
	}

	// 检查slug在项目内是否已存在
	if docVersions.ExistsBySlugAndProjectId(params.Slug, params.ProjectId) {
		return component.FailResponse("该项目下版本标识符已存在")
	}

	// 如果设置为默认版本，需要先取消其他默认版本
	if params.IsDefault == 1 {
		docVersions.ClearDefaultByProjectId(params.ProjectId)
	}

	// 创建版本实体
	version := docVersions.Entity{
		ProjectId:   params.ProjectId,
		Name:        params.Name,
		Slug:        params.Slug,
		Description: params.Description,
		Status:      params.Status,
		IsDefault:   params.IsDefault,
		SortOrder:   params.SortOrder,
	}

	// 保存版本
	rowsAffected := docVersions.SaveOrCreateById(&version)
	if rowsAffected <= 0 {
		return component.FailResponse("创建版本失败")
	}

	// 记录操作日志
	optlogger.UserOpt(req.UserId, optlogger.CreateDocVersion, version.Id, "创建文档版本: "+version.Name)

	response := DocsVersionItem{
		Id:                 version.Id,
		ProjectId:          version.ProjectId,
		ProjectName:        project.Name,
		Name:               version.Name,
		Slug:               version.Slug,
		Description:        version.Description,
		Status:             version.Status,
		IsDefault:          version.IsDefault,
		SortOrder:          version.SortOrder,
		DirectoryStructure: version.Directory,
		CreatedAt:          version.CreatedAt.Format(time.DateTime),
		UpdatedAt:          version.UpdatedAt.Format(time.DateTime),
	}

	return component.SuccessResponse(response)
}

// AdminDocsVersionUpdate 更新版本
func AdminDocsVersionUpdate(req component.BetterRequest[DocsVersionUpdateReq]) component.Response {
	params := req.Params

	// 获取原版本信息
	originalVersion := docVersions.Get(params.Id)
	if originalVersion.Id == 0 {
		return component.FailResponse("版本不存在")
	}

	// 检查项目是否存在
	project := docProjects.Get(params.ProjectId)
	if project.Id == 0 {
		return component.FailResponse("项目不存在")
	}

	// 检查slug在项目内是否被其他版本使用
	if params.Slug != originalVersion.Slug && docVersions.ExistsBySlugAndProjectIdExcludeId(params.Slug, params.ProjectId, params.Id) {
		return component.FailResponse("该项目下版本标识符已存在")
	}

	// 如果设置为默认版本，需要先取消其他默认版本
	if params.IsDefault == 1 && originalVersion.IsDefault != 1 {
		docVersions.ClearDefaultByProjectId(params.ProjectId)
	}

	// 更新版本信息
	originalVersion.ProjectId = params.ProjectId
	originalVersion.Name = params.Name
	originalVersion.Slug = params.Slug
	originalVersion.Description = params.Description
	originalVersion.Status = params.Status
	originalVersion.IsDefault = params.IsDefault
	originalVersion.SortOrder = params.SortOrder

	// 保存更新
	rowsAffected := docVersions.SaveOrCreateById(&originalVersion)
	if rowsAffected <= 0 {
		return component.FailResponse("更新版本失败")
	}

	// 记录操作日志
	optlogger.UserOpt(req.UserId, optlogger.UpdateDocVersion, originalVersion.Id, "更新文档版本: "+originalVersion.Name)

	response := DocsVersionItem{
		Id:                 originalVersion.Id,
		ProjectId:          originalVersion.ProjectId,
		ProjectName:        project.Name,
		Name:               originalVersion.Name,
		Slug:               originalVersion.Slug,
		Description:        originalVersion.Description,
		Status:             originalVersion.Status,
		IsDefault:          originalVersion.IsDefault,
		SortOrder:          originalVersion.SortOrder,
		DirectoryStructure: originalVersion.Directory,
		CreatedAt:          originalVersion.CreatedAt.Format(time.DateTime),
		UpdatedAt:          originalVersion.UpdatedAt.Format(time.DateTime),
	}

	return component.SuccessResponse(response)
}

// AdminDocsVersionDelete 删除版本（软删除）
func AdminDocsVersionDelete(req component.BetterRequest[DocsVersionDeleteReq]) component.Response {
	id := cast.ToUint64(req.GinContext.Param("id"))
	if id == 0 {
		return component.FailResponse("版本ID不能为空")
	}
	// 获取版本信息
	version := docVersions.Get(id)
	if version.Id == 0 {
		return component.FailResponse("版本不存在")
	}

	// 检查是否为默认版本
	if version.IsDefault == 1 {
		return component.FailResponse("不能删除默认版本，请先设置其他版本为默认版本")
	}

	// 检查版本下是否有文档
	if docVersions.HasDocuments(id) {
		return component.FailResponse("该版本下存在文档，无法删除")
	}

	// 执行软删除
	rowsAffected := docVersions.SoftDelete(id)
	if rowsAffected <= 0 {
		return component.FailResponse("删除版本失败")
	}

	// 记录操作日志
	optlogger.UserOpt(req.UserId, optlogger.DeleteDocVersion, version.Id, "删除文档版本: "+version.Name)

	return component.SuccessResponse(nil)
}

// AdminDocsVersionSetDefault 设置默认版本
func AdminDocsVersionSetDefault(req component.BetterRequest[component.Null]) component.Response {
	// 从URL参数中获取ID
	id := req.GinContext.Param("id")
	if id == "" {
		return component.FailResponse("版本ID不能为空")
	}

	// 获取版本信息
	version := docVersions.GetByIdString(id)
	if version.Id == 0 {
		return component.FailResponse("版本不存在")
	}

	// 检查版本状态
	if version.Status != 2 { // 只有已发布的版本才能设为默认
		return component.FailResponse("只有已发布的版本才能设为默认版本")
	}

	// 如果已经是默认版本，直接返回成功
	if version.IsDefault == 1 {
		return component.SuccessResponse(nil)
	}

	// 先取消该项目下的其他默认版本
	docVersions.ClearDefaultByProjectId(version.ProjectId)

	// 设置当前版本为默认
	version.IsDefault = 1
	rowsAffected := docVersions.SaveOrCreateById(&version)
	if rowsAffected <= 0 {
		return component.FailResponse("设置默认版本失败")
	}

	// 记录操作日志
	optlogger.UserOpt(req.UserId, optlogger.UpdateDocVersion, version.Id, "设置默认版本: "+version.Name)

	return component.SuccessResponse(nil)
}

// AdminDocsVersionDirectoryUpdate 更新版本目录结构
func AdminDocsVersionDirectoryUpdate(req component.BetterRequest[DocsVersionDirectoryUpdateReq]) component.Response {
	params := req.Params

	// 从URL参数中获取ID
	id := req.GinContext.Param("id")
	if id == "" {
		return component.FailResponse("版本ID不能为空")
	}

	// 获取版本信息
	version := docVersions.GetByIdString(id)
	if version.Id == 0 {
		return component.FailResponse("版本不存在")
	}

	// 更新目录结构
	if len(params.DirectoryStructure) == 0 {
		return component.FailResponse("空目录")
	}
	version.Directory = params.DirectoryStructure
	rowsAffected := docVersions.SaveOrCreateById(&version)
	if rowsAffected <= 0 {
		return component.FailResponse("更新目录结构失败")
	}

	// 记录操作日志
	optlogger.UserOpt(req.UserId, optlogger.UpdateDocVersion, version.Id, "更新目录结构: "+version.Name)

	return component.SuccessResponse(nil)
}
