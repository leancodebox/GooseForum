package viewrender

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/Masterminds/sprig/v3"
)

type TemplateRegistry struct {
	templates map[string]*template.Template
}

func NewRegistry(fileSystem fs.FS) (*TemplateRegistry, error) {
	r := &TemplateRegistry{
		templates: make(map[string]*template.Template),
	}

	// 1. Get Base Template (Layouts + Components)
	// Create base template with functions
	tmpl := template.New("resource_v2").
		Funcs(TemplateFuncs).
		Funcs(sprig.FuncMap())

	// Load Base Templates (Layouts and Components)
	baseTmpl := template.Must(tmpl.ParseFS(fileSystem, "templates/base/**/*.gohtml"))

	// 2. Walk and Load Views
	err := fs.WalkDir(fileSystem, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".gohtml") {
			return nil
		}

		// Skip layout and components directories
		normalizedPath := filepath.ToSlash(path)
		if strings.Contains(normalizedPath, "templates/base") {
			return nil
		}

		// Clone Base
		tmpl, err := baseTmpl.Clone()
		if err != nil {
			return err
		}

		// Parse View
		_, err = tmpl.ParseFS(fileSystem, path)
		if err != nil {
			return err
		}

		name := d.Name()
		r.templates[name] = tmpl
		return nil
	})

	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *TemplateRegistry) Render(w io.Writer, name string, data any) error {
	tmpl, ok := r.templates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}
	return tmpl.ExecuteTemplate(w, name, data)
}
