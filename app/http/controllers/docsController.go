package controllers

import (
	"fmt"
	array "github.com/leancodebox/GooseForum/app/bundles/collectionopt"
	"github.com/leancodebox/GooseForum/app/models/docs/docContents"
	"github.com/leancodebox/GooseForum/app/models/docs/docProjects"
	"github.com/leancodebox/GooseForum/app/models/docs/docVersions"
	"github.com/spf13/cast"
	"html/template"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/app/bundles/setting"
	"github.com/leancodebox/GooseForum/app/http/controllers/markdown2html"
	"github.com/leancodebox/GooseForum/app/http/controllers/viewrender"
)

// DocProject 文档项目结构
type DocProject struct {
	ID          uint64    `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      int       `json:"status"` // 1:已发布 0:草稿
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DocVersion 文档版本结构
type DocVersion struct {
	ID          uint64                      `json:"id"`
	ProjectID   uint64                      `json:"project_id"`
	Version     string                      `json:"version"`
	Name        string                      `json:"name"`
	Description string                      `json:"description"`
	Directory   []docVersions.DirectoryItem `json:"directory"` // JSON格式的目录结构
	IsDefault   bool                        `json:"is_default"`
	Status      int                         `json:"status"`
	CreatedAt   time.Time                   `json:"created_at"`
	UpdatedAt   time.Time                   `json:"updated_at"`
}

// DocContent 文档内容结构
type DocContent struct {
	ID        uint64    `json:"id"`
	VersionID uint64    `json:"version_id"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	SortOrder int       `json:"sort_order"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DocsHome 文档首页 - 显示所有项目列表
func DocsHome(c *gin.Context) {
	docProjectsList := docProjects.GetAllActive()
	projects := array.Map(docProjectsList, func(t docProjects.Entity) DocProject {
		return DocProject{
			ID:          t.Id,
			Slug:        t.Slug,
			Name:        t.Name,
			Description: t.Description,
			Status:      cast.ToInt(t.Status),
			CreatedAt:   t.CreatedAt,
			UpdatedAt:   t.UpdatedAt,
		}
	})

	viewrender.Render(c, "docs-home.gohtml", map[string]any{
		"IsProduction":  setting.IsProduction(),
		"User":          GetLoginUser(c),
		"Title":         "文档中心 - GooseForum",
		"Description":   "GooseForum 文档中心，包含完整的使用指南和API参考",
		"Projects":      projects,
		"CanonicalHref": buildCanonicalHref(c),
	})
}

// DocsVersion 版本首页 - 显示版本的第一个文档
func DocsVersion(c *gin.Context) {
	projectSlug := c.Param("project")
	versionSlug := c.Param("version")

	if projectSlug == "" {
		errorPage(c, "参数错误", "项目或版本标识不能为空")
		return
	}

	docProjectEntity := docProjects.GetBySlug(projectSlug)
	if docProjectEntity.Id == 0 {
		errorPage(c, "参数错误", "项目或版本标识不能为空")
		return
	}

	// 查找项目和版本
	var project *DocProject
	var version *DocVersion

	// 获取该项目的所有版本，用于版本切换
	projectVersions := make([]DocVersion, 0)
	docVersionList := docVersions.GetVersionByProject(docProjectEntity.Id)
	var targetVersion *docVersions.Entity
	for _, v := range docVersionList {
		if versionSlug != "" && v.Slug == versionSlug {
			targetVersion = v
			break
		} else if versionSlug == "" && v.IsDefault == 1 {
			targetVersion = v
			break
		}
		projectVersions = append(projectVersions, DocVersion{
			ID:          v.Id,
			ProjectID:   v.ProjectId,
			Version:     v.Slug,
			Name:        v.Description,
			Description: v.Description,
			Directory:   v.Directory,
			IsDefault:   v.IsDefault == 1,
			Status:      cast.ToInt(v.Status),
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		})
	}

	if targetVersion == nil {
		errorPage(c, "版本不存在", "找不到指定的文档版本")
		return
	}

	project = &DocProject{
		ID:          docProjectEntity.Id,
		Slug:        docProjectEntity.Slug,
		Name:        docProjectEntity.Name,
		Description: docProjectEntity.Description,
		Status:      cast.ToInt(docProjectEntity.Status),
		CreatedAt:   docProjectEntity.CreatedAt,
		UpdatedAt:   docProjectEntity.UpdatedAt,
	}
	version = &DocVersion{
		ID:          targetVersion.Id,
		ProjectID:   targetVersion.ProjectId,
		Version:     targetVersion.Slug,
		Name:        targetVersion.Description,
		Description: targetVersion.Description,
		Directory:   targetVersion.Directory,
		IsDefault:   targetVersion.IsDefault == 1,
		Status:      cast.ToInt(targetVersion.Status),
		CreatedAt:   targetVersion.CreatedAt,
		UpdatedAt:   targetVersion.UpdatedAt,
	}

	// 构建面包屑导航
	breadcrumbs := []map[string]string{
		{"title": "文档中心", "url": "/docs"},
		{"title": project.Name, "url": ""},
	}

	viewrender.Render(c, "docs-version.gohtml", map[string]any{
		"IsProduction":    setting.IsProduction(),
		"User":            GetLoginUser(c),
		"Title":           fmt.Sprintf("%s %s - %s - GooseForum", project.Name, version.Name, "文档中心"),
		"Description":     fmt.Sprintf("%s - %s", project.Description, version.Description),
		"Project":         project,
		"Version":         version,
		"Directory":       version.Directory,
		"ProjectVersions": projectVersions,
		"Breadcrumbs":     breadcrumbs,
		"CanonicalHref":   buildCanonicalHref(c),
	})
}

// DocsContent 文档内容页面
func DocsContent(c *gin.Context) {
	projectSlug := c.Param("project")
	versionSlug := c.Param("version")
	contentSlug := c.Param("content")

	if projectSlug == "" || versionSlug == "" || contentSlug == "" {
		errorPage(c, "参数错误", "项目、版本或内容标识不能为空")
		return
	}

	docProjectEntity := docProjects.GetBySlug(projectSlug)
	if docProjectEntity.Id == 0 {
		errorPage(c, "参数错误", "项目或版本标识不能为空")
		return
	}

	// 获取该项目的所有版本，用于版本切换
	projectVersions := make([]DocVersion, 0)
	docVersionList := docVersions.GetVersionByProject(docProjectEntity.Id)
	var targetVersion *docVersions.Entity
	for _, v := range docVersionList {
		if v.Slug == versionSlug {
			targetVersion = v
			break
		}
		projectVersions = append(projectVersions, DocVersion{
			ID:          v.Id,
			ProjectID:   v.ProjectId,
			Version:     v.Slug,
			Name:        v.Description,
			Description: v.Description,
			Directory:   v.Directory,
			IsDefault:   v.IsDefault == 1,
			Status:      cast.ToInt(v.Status),
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		})
	}

	if targetVersion == nil {
		errorPage(c, "版本不存在", "找不到指定的文档版本")
		return
	}
	fmt.Println(targetVersion.Id, contentSlug)
	docContentEntity := docContents.GetBySlug(targetVersion.Id, contentSlug)
	if docContentEntity.Id == 0 {
		errorPage(c, "内容不存在", "找不到指定的文档内容")
		return
	}

	// 查找项目、版本和内容
	var project *DocProject
	var version *DocVersion
	var content *DocContent

	project = &DocProject{
		ID:          docProjectEntity.Id,
		Slug:        docProjectEntity.Slug,
		Name:        docProjectEntity.Name,
		Description: docProjectEntity.Description,
		Status:      cast.ToInt(docProjectEntity.Status),
		CreatedAt:   docProjectEntity.CreatedAt,
		UpdatedAt:   docProjectEntity.UpdatedAt,
	}
	version = &DocVersion{
		ID:          targetVersion.Id,
		ProjectID:   targetVersion.ProjectId,
		Version:     targetVersion.Slug,
		Name:        targetVersion.Description,
		Description: targetVersion.Description,
		Directory:   targetVersion.Directory,
		IsDefault:   targetVersion.IsDefault == 1,
		Status:      cast.ToInt(targetVersion.Status),
		CreatedAt:   targetVersion.CreatedAt,
		UpdatedAt:   targetVersion.UpdatedAt,
	}

	content = &DocContent{
		ID:        docContentEntity.Id,
		VersionID: docContentEntity.VersionId,
		Slug:      docContentEntity.Slug,
		Title:     docContentEntity.Title,
		Content:   docContentEntity.Content,
		SortOrder: docContentEntity.SortOrder,
		Status:    1,
		CreatedAt: docContentEntity.CreatedAt,
		UpdatedAt: docContentEntity.UpdatedAt,
	}

	// 获取该版本的所有内容，用于构建导航
	versionContents := make([]DocContent, 0)

	// 构建面包屑导航
	breadcrumbs := []map[string]string{
		{"title": "文档中心", "url": "/docs"},
		{"title": project.Name, "url": fmt.Sprintf("/docs/%s", project.Slug)},
		{"title": content.Title, "url": ""},
	}

	// 渲染Markdown内容为HTML
	htmlContent := template.HTML(markdown2html.MarkdownToHTML(content.Content))

	viewrender.Render(c, "docs-content.gohtml", map[string]any{
		"IsProduction":    setting.IsProduction(),
		"User":            GetLoginUser(c),
		"Title":           fmt.Sprintf("%s - %s - %s", content.Title, project.Name, "GooseForum"),
		"Description":     fmt.Sprintf("%s - %s", content.Title, project.Description),
		"Project":         project,
		"Version":         version,
		"Content":         content,
		"HTMLContent":     htmlContent,
		"Directory":       version.Directory,
		"VersionContents": versionContents,
		"ProjectVersions": projectVersions,
		"Breadcrumbs":     breadcrumbs,
		"CanonicalHref":   buildCanonicalHref(c),
		"CurrentSlug":     contentSlug,
	})
}
