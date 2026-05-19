package forum

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/Masterminds/sprig/v3"
	"github.com/gin-gonic/gin"
	"github.com/leancodebox/GooseForum/resource"
)

type templateRegistry struct {
	templates map[string]*template.Template
}

type templateData struct {
	Payload PagePayload
	Lang    string
}

var currentRegistry = mustNewRegistry()
var currentRegistryErr error

func ReloadTemplates() {
	currentRegistry = mustNewRegistry()
}

func mustNewRegistry() *templateRegistry {
	registry, err := newRegistry(resource.GetTemplateFS())
	if err != nil {
		currentRegistryErr = err
		slog.Error("failed to load resource templates", "err", err)
	}
	return registry
}

func newRegistry(fileSystem fs.FS) (*templateRegistry, error) {
	base := template.New("goose_resource").
		Funcs(templateFuncs()).
		Funcs(sprig.FuncMap())

	sharedFiles, err := templateFiles(fileSystem, "templates/layout", "templates/partials")
	if err != nil {
		return nil, err
	}
	if _, err := base.ParseFS(fileSystem, sharedFiles...); err != nil {
		return nil, err
	}

	registry := &templateRegistry{templates: map[string]*template.Template{}}
	err = fs.WalkDir(fileSystem, "templates/pages", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".gohtml") {
			return nil
		}
		tmpl, err := base.Clone()
		if err != nil {
			return err
		}
		if _, err := tmpl.ParseFS(fileSystem, path); err != nil {
			return err
		}
		registry.templates[d.Name()] = tmpl
		return nil
	})
	if err != nil {
		return nil, err
	}
	return registry, nil
}

func templateFiles(fileSystem fs.FS, roots ...string) ([]string, error) {
	var files []string
	for _, root := range roots {
		err := fs.WalkDir(fileSystem, root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || !strings.HasSuffix(d.Name(), ".gohtml") {
				return nil
			}
			files = append(files, path)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

func (r *templateRegistry) render(w io.Writer, name string, data any) error {
	tmpl, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}
	return tmpl.ExecuteTemplate(w, name, data)
}

func renderPage(c *gin.Context, templateName string, payload PagePayload) {
	if isPageRequest(c) {
		c.JSON(http.StatusOK, payload)
		return
	}
	if currentRegistry == nil {
		if currentRegistryErr != nil {
			c.String(http.StatusInternalServerError, currentRegistryErr.Error())
			return
		}
		c.String(http.StatusInternalServerError, "resource template registry is not initialized")
		return
	}
	if err := currentRegistry.render(c.Writer, filepath.Base(templateName), templateData{
		Payload: payload,
		Lang:    requestLang(c),
	}); err != nil {
		slog.Error("render resource template failed", "template", templateName, "err", err)
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func isPageRequest(c *gin.Context) bool {
	return c.GetHeader("X-Goose-Page") == "true"
}
