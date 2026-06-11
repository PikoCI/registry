package templates

import (
	"embed"
	"text/template"
	"io/fs"
	"path/filepath"
)

//go:embed views
var templateFS embed.FS

// Templates holds all parsed templates keyed by their path (e.g. "views/layouts/index.tmpl").
var Templates map[string]*template.Template

func init() {
	Templates = make(map[string]*template.Template)
	fs.WalkDir(templateFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext == ".tmpl" || ext == ".html" {
			t, parseErr := template.ParseFS(templateFS, path)
			if parseErr != nil {
				return parseErr
			}
			Templates[path] = t
		}
		return nil
	})
}
