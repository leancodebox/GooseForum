package viewrender

import (
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/resourcev2"
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
		ht4gooseforum = resourcev2.GetTemplates()
	})
}

func Reload() {
	ht4gooseforum = resourcev2.GetTemplates()
}

func init() {
	getHt4gooseforum()
}

func Render(c *gin.Context, name string, data any) {
	if err := ht4gooseforum.ExecuteTemplate(c.Writer, name, data); err != nil {
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
