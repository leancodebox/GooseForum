package viewrender

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/resource"
	"html/template"
	"log/slog"
	"net/http"
	"strings"
	"sync"
)

var ht4gooseforum *template.Template
var htht4gooseforumOnce sync.Once

func getHt4gooseforum() {
	htht4gooseforumOnce.Do(func() {
		// 创建基础模板
		ht4gooseforum = resource.GetTemplates()
	})
}

func Reload() {
	ht4gooseforum = resource.GetTemplates()
}

func init() {
	getHt4gooseforum()
}

func Render(c *gin.Context, name string, data any) {
	// 从cookie中读取主题设置
	theme := "light" // 默认主题
	if themeCookie, err := c.Cookie("theme"); err == nil && themeCookie != "" {
		theme = themeCookie
	}
	// 将数据转换为map并添加主题信息
	var templateData map[string]any
	if dataMap, ok := data.(map[string]any); ok {
		templateData = dataMap
	} else {
		templateData = map[string]any{"Data": data}
	}
	templateData["Theme"] = theme
	if err := ht4gooseforum.ExecuteTemplate(c.Writer, name, templateData); err != nil {
		slog.Error("render template err", "err", err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
	}
}

// generateTemplateName 将文件路径转换为唯一的模板名
func generateTemplateName(path string) string {
	// 移除前导的 ./ 或 .
	path = strings.TrimPrefix(path, "./")
	path = strings.TrimPrefix(path, ".")
	// 替换路径分隔符为下划线
	name := strings.ReplaceAll(path, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	// 移除 .html 扩展名
	name = strings.TrimSuffix(name, ".html")
	return name
}
