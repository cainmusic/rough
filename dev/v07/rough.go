package rough

import (
	"html/template"
	"log"
	"net/http"

	"github.com/cainmusic/rough/dev/v07/render"
)

type HandleFunc func(*Context)

const defaultMultipartMemory = 32 << 20

type Engine struct {
	RouterGroup
	maps map[string]map[string][]HandleFunc

	delims     render.Delims
	HTMLRender render.HTMLRender
	FuncMap    template.FuncMap
}

func New() *Engine {
	en := &Engine{
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "/",
		},
		maps: make(map[string]map[string][]HandleFunc),

		FuncMap: template.FuncMap{},
		delims:  render.Delims{Left: "{{", Right: "}}"},
	}
	en.RouterGroup.engine = en
	return en
}

func (en *Engine) Use(fs ...HandleFunc) {
	en.RouterGroup.Use(fs...)
}

func (en *Engine) LoadHTMLGlob(pattern string) {
	templ := template.Must(
		template.New("").
			Delims(en.delims.Left, en.delims.Right).
			Funcs(en.FuncMap).
			ParseGlob(pattern))
	en.SetHTMLTemplate(templ)
}

func (en *Engine) LoadHTMLFiles(files ...string) {
	templ := template.Must(
		template.New("").
			Delims(en.delims.Left, en.delims.Right).
			Funcs(en.FuncMap).
			ParseFiles(files...))
	en.SetHTMLTemplate(templ)
}

func (en *Engine) SetHTMLTemplate(templ *template.Template) {
	if len(en.maps) > 0 {
		log.Println("need to set html template before set route")
	}

	en.HTMLRender = render.HTMLProduction{Template: templ.Funcs(en.FuncMap)}
}

func (en *Engine) SetFuncMap(funcMap template.FuncMap) {
	en.FuncMap = funcMap
}

func (en *Engine) addRoute(method, path string, handlers []HandleFunc) {
	rMap, ok := en.maps[method]
	if !ok {
		en.maps[method] = make(map[string][]HandleFunc)
		rMap = en.maps[method]
	}
	rMap[path] = handlers
}

func (en *Engine) getRoute(method, path string) []HandleFunc {
	rMap, ok := en.maps[method]
	if !ok {
		return nil
	}
	handlers, ok := rMap[path]
	if !ok {
		return nil
	}
	return handlers
}

func (en *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := &Context{
		engine: en,
		W:      w,
		R:      r,
		index:  -1,
	}
	c.handlers = en.getRoute(c.R.Method, c.R.URL.Path)
	log.Println(c.R.Method, c.R.URL.Path, len(c.handlers), "handler(s)")
	if c.handlers == nil {
		log.Println("request", c.R.Method, c.R.URL.Path, "not found")
		c.String(http.StatusNotFound, "Not Found")
		return
	}
	en.handleRequest(c)
}

func (en *Engine) handleRequest(c *Context) {
	c.Next()
}

func (en *Engine) Run() {
	en.debugPrintMap()

	http.ListenAndServe(":8888", en)
}

func (en *Engine) debugPrintMap() {
	for method, rMap := range en.maps {
		log.Println("method", method)
		for url, handlers := range rMap {
			log.Println(url, len(handlers), "handler(s)")
		}
	}
}
