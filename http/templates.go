package http

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

//go:embed views/*.tmpl
var templateFS embed.FS

// TemplateCache caches the available templates.
type TemplateCache map[string]*template.Template

// NewTemplateCache instantiates the template store with the embedded
// template file system.
func NewTemplateCache() (TemplateCache, error) {
	cache, err := cacheTemplates(templateFS, "views")
	if err != nil {
		return nil, errors.Wrap(err, "template cache initialiation failed")
	}
	return cache, nil
}

func (c TemplateCache) Render(w io.Writer, name string, data interface{}) error {
	tmpl, ok := c[name]
	if !ok {
		return fmt.Errorf("template does not exist: %v", name)
	}
	return tmpl.Execute(w, data)
}

// cacheTemplates caches templates in a filesystem and direction.
func cacheTemplates(fsys fs.FS, dir string) (TemplateCache, error) {
	cache := map[string]*template.Template{}

	pages, err := fs.Glob(fsys, filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := strings.TrimSuffix(filepath.Base(page), ".page.tmpl")

		ts, err := template.ParseFS(fsys, page)
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFS(fsys, filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFS(fsys, filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}
