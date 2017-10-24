package pathfork

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"bitbucket.org/jtyburke/pathfork/app/config"
	"bitbucket.org/jtyburke/pathfork/app/forms"
	"bitbucket.org/jtyburke/pathfork/app/pages"
)

/////// utility functions on all templates for rendering
//
// URLFor returns URL of name based on AppRoutes. Bootstrapped at runtime
func URLFor(name string, args ...string) string {
	return namePathMap[name]
}

// StaticURL simply prepends the static URL to the path
func StaticURL(path string) string {
	return fmt.Sprintf("%s%s", StaticRoute, path)
}

func WrapField(field forms.FormField) template.HTML {
	rendered := fmt.Sprintf(`<div class="row"><div class="col-md-9">%v</div></div>`, field.Render())
	return template.HTML(rendered)
}

func WrapTextAreaField(field forms.FormField, rows, cols string) template.HTML {
	rendered := fmt.Sprintf(
		`<div class="row"><div class="col-md-%v">%v</div></div>`,
		cols, field.Render(rows, cols),
	)
	return template.HTML(rendered)
}

func AsHTML(input string) template.HTML {
	return template.HTML(input)
}

func add(x, y int) int {
	return x + y
}

// TemplateRenderer caches the template files and adds utility functions
type TemplateRenderer struct {
	templates map[string]*template.Template
}

func NewTemplateRenderer() *TemplateRenderer {
	templates := make(map[string]*template.Template)

	templateFiles, err := filepath.Glob(config.TemplatePath + "*.html")
	if err != nil {
		panic("Could not load files in templateDir")
	}
	universalTemplates := []string{}
	for i := range templateFiles {
		filename := strings.TrimSuffix(filepath.Base(templateFiles[i]), ".html")
		if filename[0] == '_' {
			universalTemplates = append(universalTemplates, templateFiles[i])
		}
	}
	for _, file := range templateFiles {
		key := strings.TrimSuffix(filepath.Base(file), ".html")
		newTmpl := &template.Template{}
		if key == "work_export" {
			newTmpl = template.New("base").Funcs(template.FuncMap{
				"URLFor":            URLFor,
				"StaticURL":         StaticURL,
				"WrapField":         WrapField,
				"WrapTextAreaField": WrapTextAreaField,
				"AsHTML":            AsHTML,
				"Add":               add,
			})
			fileStr, err := ioutil.ReadFile(file)
			if err != nil {
				panic("Could not load work export file")
			}
			//TODO FIXME what the hell, why doesn't ParseFiles work here?
			newTmpl.Parse(string(fileStr))
		} else {
			newTmpl = template.New("base").Funcs(template.FuncMap{
				"URLFor":            URLFor,
				"StaticURL":         StaticURL,
				"WrapField":         WrapField,
				"WrapTextAreaField": WrapTextAreaField,
				"AsHTML":            AsHTML,
			})
			newTmpl.ParseFiles(append(universalTemplates, file)...)
		}
		templates[key] = newTmpl
	}
	return &TemplateRenderer{templates: templates}
}

func (tr *TemplateRenderer) RenderPage(w http.ResponseWriter, tmplName string, webpage pages.WebPage) error {
	return renderTemplate(w, tr, tmplName, webpage)
}

func renderTemplate(w http.ResponseWriter, tr *TemplateRenderer, tmplName string, webpage pages.WebPage) error {
	tmpl, ok := tr.templates[tmplName]
	if !ok {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(
			"<h2>Better tell the administrator something went wrong with the template.</h2>"))
		return fmt.Errorf("Could not locate template %v", tmplName)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if tmplName == "work_export" {
		return tmpl.ExecuteTemplate(w, "base", webpage)
	} else {
		return tmpl.ExecuteTemplate(w, "base", webpage)
	}
}
