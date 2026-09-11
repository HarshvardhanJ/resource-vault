package web

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"strings"
)

type Renderer struct {
	templates map[string]*template.Template
}

func NewRenderer(templateFS fs.FS) (*Renderer, error) {
	funcMap := template.FuncMap{
		"formatYear": func(year int) string {
			if year <= 0 {
				return ""
			}
			return fmt.Sprintf("%d-%02d", year, (year+1)%100)
		},
		"formatBytes": func(size int64) string {
			const unit = 1024
			if size < unit {
				return fmt.Sprintf("%d B", size)
			}
			div, exp := int64(unit), 0
			for n := size / unit; n >= unit; n /= unit {
				div *= unit
				exp++
			}
			units := []string{"KB", "MB", "GB", "TB"}
			return fmt.Sprintf("%.1f %s", float64(size)/float64(div), units[exp])
		},
		"upper": strings.ToUpper,
		"lower": strings.ToLower,
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, fmt.Errorf("dict requires even number of arguments")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, fmt.Errorf("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
	}

	r := &Renderer{
		templates: make(map[string]*template.Template),
	}

	// Read all files in pages and partials
	pages, err := fs.Glob(templateFS, "pages/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to glob pages: %w", err)
	}

	partials, err := fs.Glob(templateFS, "partials/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to glob partials: %w", err)
	}

	layoutFiles, err := fs.Glob(templateFS, "layouts/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to glob layouts: %w", err)
	}

	// Parse complete pages with layouts and partials
	for _, page := range pages {
		pageName := strings.TrimPrefix(page, "pages/")
		files := append([]string{}, layoutFiles...)
		files = append(files, partials...)
		files = append(files, page)

		tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS, files...)
		if err != nil {
			return nil, fmt.Errorf("failed to parse template %s: %w", page, err)
		}
		r.templates[pageName] = tmpl
	}

	// Parse individual partials for HTMX fragment responses
	for _, partial := range partials {
		partialName := strings.TrimPrefix(partial, "partials/")
		files := append([]string{partial}, partials...)

		tmpl, err := template.New("").Funcs(funcMap).ParseFS(templateFS, files...)
		if err != nil {
			return nil, fmt.Errorf("failed to parse partial %s: %w", partial, err)
		}
		r.templates[partialName] = tmpl
	}

	return r, nil
}

func (r *Renderer) Render(w http.ResponseWriter, name string, data interface{}) {
	tmpl, ok := r.templates[name]
	if !ok {
		http.Error(w, fmt.Sprintf("Template %s not found", name), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	// Determine root template name
	rootName := "base.html"
	if !strings.HasSuffix(name, ".html") {
		name += ".html"
	}
	// For standalone partials without layout
	if strings.HasPrefix(name, "partials/") || tmpl.Lookup(rootName) == nil {
		rootName = name
		if strings.HasPrefix(name, "partials/") {
			rootName = strings.TrimPrefix(name, "partials/")
		}
	}

	err := tmpl.ExecuteTemplate(&buf, rootName, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Template execution error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.Copy(w, &buf)
}

func (r *Renderer) RenderPartial(w http.ResponseWriter, name string, data interface{}) {
	tmpl, ok := r.templates[name]
	if !ok {
		http.Error(w, fmt.Sprintf("Partial %s not found", name), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, name, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Partial execution error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.Copy(w, &buf)
}
