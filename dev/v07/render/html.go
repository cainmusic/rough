package render

import (
	"html/template"
	"net/http"
)

type Delims struct {
	Left  string
	Right string
}

type HTMLRender interface {
	Instance(string, any) Render
}

type HTMLProduction struct {
	Template *template.Template
	Delims   Delims
}

func (r HTMLProduction) Instance(name string, data any) Render {
	return HTML{
		Template: r.Template,
		Name:     name,
		Data:     data,
	}
}

type HTML struct {
	Template *template.Template
	Name     string
	Data     any
}

var htmlContentType = []string{"text/html; charset=utf-8"}

func (r HTML) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	if r.Name == "" {
		return r.Template.Execute(w, r.Data)
	}
	return r.Template.ExecuteTemplate(w, r.Name, r.Data)
}

func (r HTML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, htmlContentType)
}
