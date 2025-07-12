package controllers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
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
	ID          uint64    `json:"id"`
	ProjectID   uint64    `json:"project_id"`
	Version     string    `json:"version"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Directory   string    `json:"directory"` // JSON格式的目录结构
	IsDefault   bool      `json:"is_default"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DocContent 文档内容结构
type DocContent struct {
	ID        uint64    `json:"id"`
	VersionID uint64    `json:"version_id"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	ParentID  uint64    `json:"parent_id"`
	SortOrder int       `json:"sort_order"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DirectoryItem 目录项结构
type DirectoryItem struct {
	Title       string           `json:"title"`
	Slug        string           `json:"slug"`
	Description string           `json:"description,omitempty"`
	Children    []*DirectoryItem `json:"children,omitempty"`
}

// 从JSON文件加载数据的函数
func loadMockData() ([]DocProject, []DocVersion, []DocContent, error) {
	// 从JSON文件读取模拟数据
	data, err := os.ReadFile("mock_docs_data.json")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("读取模拟数据文件失败: %v", err)
	}

	var mockData struct {
		Projects []struct {
			ID          int       `json:"id"`
			Slug        string    `json:"slug"`
			Name        string    `json:"name"`
			Description string    `json:"description"`
			Status      int       `json:"status"`
			CreatedAt   time.Time `json:"created_at"`
			UpdatedAt   time.Time `json:"updated_at"`
		} `json:"projects"`
		Versions []struct {
			ID          int       `json:"id"`
			ProjectID   int       `json:"project_id"`
			Version     string    `json:"version"`
			Name        string    `json:"name"`
			Description string    `json:"description"`
			Directory   string    `json:"directory"`
			IsDefault   bool      `json:"is_default"`
			Status      int       `json:"status"`
			CreatedAt   time.Time `json:"created_at"`
			UpdatedAt   time.Time `json:"updated_at"`
		} `json:"versions"`
		Contents []struct {
			ID        int       `json:"id"`
			VersionID int       `json:"version_id"`
			Slug      string    `json:"slug"`
			Title     string    `json:"title"`
			Content   string    `json:"content"`
			ParentID  int       `json:"parent_id"`
			SortOrder int       `json:"sort_order"`
			Status    int       `json:"status"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"contents"`
	}

	if err := json.Unmarshal(data, &mockData); err != nil {
		return nil, nil, nil, fmt.Errorf("解析模拟数据失败: %v", err)
	}

	// 转换项目数据
	var projects []DocProject
	for _, p := range mockData.Projects {
		projects = append(projects, DocProject{
			ID:          uint64(p.ID),
			Slug:        p.Slug,
			Name:        p.Name,
			Description: p.Description,
			Status:      p.Status,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		})
	}

	// 转换版本数据
	var versions []DocVersion
	for _, v := range mockData.Versions {
		versions = append(versions, DocVersion{
			ID:          uint64(v.ID),
			ProjectID:   uint64(v.ProjectID),
			Version:     v.Version,
			Name:        v.Name,
			Description: v.Description,
			Directory:   v.Directory,
			IsDefault:   v.IsDefault,
			Status:      v.Status,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		})
	}

	// 转换内容数据
	var contents []DocContent
	for _, c := range mockData.Contents {
		contents = append(contents, DocContent{
			ID:        uint64(c.ID),
			VersionID: uint64(c.VersionID),
			Slug:      c.Slug,
			Title:     c.Title,
			Content:   c.Content,
			ParentID:  uint64(c.ParentID),
			SortOrder: c.SortOrder,
			Status:    c.Status,
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		})
	}

	return projects, versions, contents, nil
}

// DocsHome 文档首页 - 显示所有项目列表
func DocsHome(c *gin.Context) {
	projects, _, _, err := loadMockData()
	if err != nil {
		errorPage(c, "加载数据失败", "无法加载文档数据")
		return
	}

	// 过滤已发布的项目
	publishedProjects := make([]DocProject, 0)
	for _, project := range projects {
		if project.Status == 1 {
			publishedProjects = append(publishedProjects, project)
		}
	}

	viewrender.Render(c, "docs-home.gohtml", map[string]any{
		"IsProduction":  setting.IsProduction(),
		"User":          GetLoginUser(c),
		"Title":         "文档中心 - GooseForum",
		"Description":   "GooseForum 文档中心，包含完整的使用指南和API参考",
		"Projects":      publishedProjects,
		"CanonicalHref": buildCanonicalHref(c),
	})
}

// DocsProject 项目首页 - 显示默认版本的文档
func DocsProject(c *gin.Context) {
	projectSlug := c.Param("project")
	if projectSlug == "" {
		errorPage(c, "项目不存在", "项目标识不能为空")
		return
	}

	projects, versions, _, err := loadMockData()
	if err != nil {
		errorPage(c, "加载数据失败", "无法加载文档数据")
		return
	}

	// 查找项目
	var project *DocProject
	for _, p := range projects {
		if p.Slug == projectSlug && p.Status == 1 {
			project = &p
			break
		}
	}

	if project == nil {
		errorPage(c, "项目不存在", "找不到指定的文档项目")
		return
	}

	// 获取该项目的所有版本
	projectVersions := make([]DocVersion, 0)
	for _, v := range versions {
		if v.ProjectID == project.ID && v.Status == 1 {
			projectVersions = append(projectVersions, v)
		}
	}

	// 构建面包屑导航
	breadcrumbs := []map[string]string{
		{"title": "文档中心", "url": "/docs"},
		{"title": project.Name, "url": ""},
	}

	viewrender.Render(c, "docs-project.gohtml", map[string]any{
		"IsProduction":  setting.IsProduction(),
		"User":          GetLoginUser(c),
		"Title":         fmt.Sprintf("%s - 文档中心 - GooseForum", project.Name),
		"Description":   project.Description,
		"Project":       project,
		"Versions":      projectVersions,
		"Breadcrumbs":   breadcrumbs,
		"CanonicalHref": buildCanonicalHref(c),
	})
}

// DocsVersion 版本首页 - 显示版本的第一个文档
func DocsVersion(c *gin.Context) {
	projectSlug := c.Param("project")
	versionSlug := c.Param("version")

	if projectSlug == "" || versionSlug == "" {
		errorPage(c, "参数错误", "项目或版本标识不能为空")
		return
	}

	projects, versions, _, err := loadMockData()
	if err != nil {
		errorPage(c, "加载数据失败", "无法加载文档数据")
		return
	}

	// 查找项目和版本
	var project *DocProject
	var version *DocVersion

	for _, p := range projects {
		if p.Slug == projectSlug && p.Status == 1 {
			project = &p
			break
		}
	}

	if project == nil {
		errorPage(c, "项目不存在", "找不到指定的文档项目")
		return
	}

	for _, v := range versions {
		if v.ProjectID == project.ID && v.Version == versionSlug && v.Status == 1 {
			version = &v
			break
		}
	}

	if version == nil {
		errorPage(c, "版本不存在", "找不到指定的文档版本")
		return
	}

	// 解析目录结构
	var directory []DirectoryItem
	if err := json.Unmarshal([]byte(version.Directory), &directory); err != nil {
		errorPage(c, "数据错误", "无法解析目录结构")
		return
	}

	// 获取该项目的所有版本，用于版本切换
	projectVersions := make([]DocVersion, 0)
	for _, v := range versions {
		if v.ProjectID == project.ID && v.Status == 1 {
			projectVersions = append(projectVersions, v)
		}
	}

	// 构建面包屑导航
	breadcrumbs := []map[string]string{
		{"title": "文档中心", "url": "/docs"},
		{"title": project.Name, "url": fmt.Sprintf("/docs/%s", project.Slug)},
		{"title": version.Name, "url": ""},
	}

	viewrender.Render(c, "docs-version.gohtml", map[string]any{
		"IsProduction":    setting.IsProduction(),
		"User":            GetLoginUser(c),
		"Title":           fmt.Sprintf("%s %s - %s - GooseForum", project.Name, version.Name, "文档中心"),
		"Description":     fmt.Sprintf("%s - %s", project.Description, version.Description),
		"Project":         project,
		"Version":         version,
		"Directory":       directory,
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

	projects, versions, contents, err := loadMockData()
	if err != nil {
		errorPage(c, "加载数据失败", "无法加载文档数据")
		return
	}

	// 查找项目、版本和内容
	var project *DocProject
	var version *DocVersion
	var content *DocContent

	for _, p := range projects {
		if p.Slug == projectSlug && p.Status == 1 {
			project = &p
			break
		}
	}

	if project == nil {
		errorPage(c, "项目不存在", "找不到指定的文档项目")
		return
	}

	for _, v := range versions {
		if v.ProjectID == project.ID && v.Version == versionSlug && v.Status == 1 {
			version = &v
			break
		}
	}

	if version == nil {
		errorPage(c, "版本不存在", "找不到指定的文档版本")
		return
	}

	for _, cnt := range contents {
		if cnt.VersionID == version.ID && cnt.Slug == contentSlug && cnt.Status == 1 {
			content = &cnt
			break
		}
	}

	if content == nil {
		errorPage(c, "内容不存在", "找不到指定的文档内容")
		return
	}

	// 解析目录结构
	var directory []DirectoryItem
	if err := json.Unmarshal([]byte(version.Directory), &directory); err != nil {
		errorPage(c, "数据错误", "无法解析目录结构")
		return
	}

	// 获取该版本的所有内容，用于构建导航
	versionContents := make([]DocContent, 0)
	for _, cnt := range contents {
		if cnt.VersionID == version.ID && cnt.Status == 1 {
			versionContents = append(versionContents, cnt)
		}
	}

	// 构建面包屑导航
	breadcrumbs := []map[string]string{
		{"title": "文档中心", "url": "/docs"},
		{"title": project.Name, "url": fmt.Sprintf("/docs/%s", project.Slug)},
		{"title": version.Name, "url": fmt.Sprintf("/docs/%s/%s", project.Slug, version.Version)},
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
		"Directory":       directory,
		"VersionContents": versionContents,
		"Breadcrumbs":     breadcrumbs,
		"CanonicalHref":   buildCanonicalHref(c),
		"CurrentSlug":     contentSlug,
	})
}
