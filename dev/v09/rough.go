package rough

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"regexp"

	"github.com/cainmusic/rough/dev/v09/render"
)

type HandleFunc func(*Context)

const defaultMultipartMemory = 32 << 20

var default404Body = []byte("404 page not found")

var regSafePrefix = regexp.MustCompile("[^a-zA-Z0-9/-]+")
var regRemoveRepeatedChar = regexp.MustCompile("/{2,}")

type Engine struct {
	RouterGroup
	trees methodTrees

	RedirectTrailingSlash bool
	RedirectFixedPath     bool

	delims     render.Delims
	HTMLRender render.HTMLRender
	FuncMap    template.FuncMap

	maxParams   uint16
	maxSections uint16

	noRoute    []HandleFunc
	allNoRoute []HandleFunc
}

type RouteInfo struct {
	Method      string
	Path        string
	Handler     string
	HandlerFunc HandleFunc
	HandlerLen  int
}

type RoutesInfo []RouteInfo

func New() *Engine {
	en := &Engine{
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "/",
		},
		trees: make(methodTrees, 0, 9),

		RedirectTrailingSlash: true,
		RedirectFixedPath:     false,

		FuncMap: template.FuncMap{},
		delims:  render.Delims{Left: "{{", Right: "}}"},
	}
	en.RouterGroup.engine = en
	return en
}

func (en *Engine) Use(fs ...HandleFunc) {
	en.RouterGroup.Use(fs...)
	en.rebuild404Handlers()
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
	en.HTMLRender = render.HTMLProduction{Template: templ.Funcs(en.FuncMap)}
}

func (en *Engine) SetFuncMap(funcMap template.FuncMap) {
	en.FuncMap = funcMap
}

func (en *Engine) NoRoute(handlers ...HandleFunc) {
	en.noRoute = handlers
	en.rebuild404Handlers()
}

func (en *Engine) rebuild404Handlers() {
	en.allNoRoute = en.combineHandlers(en.noRoute)
}

func (en *Engine) addRoute(method, path string, handlers []HandleFunc) {
	assert1(path[0] == '/', "path must begin with '/'")
	assert1(method != "", "HTTP method can not be empty")
	assert1(len(handlers) > 0, "there must be at least one handler")

	root := en.trees.get(method)
	if root == nil {
		root = new(node)
		root.fullPath = "/"
		en.trees = append(en.trees, methodTree{method: method, root: root})
	}
	root.addRoute(path, handlers)

	if paramsCount := countParams(path); paramsCount > en.maxParams {
		en.maxParams = paramsCount
	}

	if sectionsCount := countSections(path); sectionsCount > en.maxSections {
		en.maxSections = sectionsCount
	}
}

func (engine *Engine) Routes() (routes RoutesInfo) {
	for _, tree := range engine.trees {
		routes = iterate("", tree.method, routes, tree.root)
	}
	return routes
}

func iterate(path, method string, routes RoutesInfo, root *node) RoutesInfo {
	path += root.path
	if len(root.handlers) > 0 {
		handlerFunc := root.handlers[len(root.handlers)-1]
		routes = append(routes, RouteInfo{
			Method:      method,
			Path:        path,
			Handler:     nameOfFunction(handlerFunc),
			HandlerFunc: handlerFunc,
			HandlerLen:  len(root.handlers),
		})
	}
	for _, child := range root.children {
		routes = iterate(path, method, routes, child)
	}
	return routes
}

func (en *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	params := make(Params, 0, en.maxParams)
	skippedNodes := make([]skippedNode, 0, en.maxSections)
	c := &Context{
		engine:       en,
		W:            w,
		R:            r,
		params:       &params,
		skippedNodes: &skippedNodes,
		index:        -1,
	}

	en.handleRequest(c)
}

func (en *Engine) handleRequest(c *Context) {
	httpMethod := c.R.Method
	rPath := c.R.URL.Path

	t := en.trees
	for i := 0; i < len(t); i++ {
		if t[i].method != httpMethod {
			continue
		}
		root := t[i].root

		value := root.getValue(rPath, c.params, c.skippedNodes, false)
		if value.params != nil {
			c.Params = *value.params
		}
		if value.handlers != nil {
			c.handlers = value.handlers
			c.fullPath = value.fullPath
			log.Println(httpMethod, c.fullPath, len(c.handlers), "handler[s]")
			c.Next()
			return
		}
		if httpMethod != http.MethodConnect && rPath != "/" {
			if value.tsr && en.RedirectTrailingSlash {
				redirectTrailingSlash(c)
				return
			}
			if en.RedirectFixedPath && redirectFixedPath(c, root, en.RedirectFixedPath) {
				return
			}
		}
		break
	}
	c.handlers = en.allNoRoute
	c.String(http.StatusNotFound, string(default404Body))
}

func (en *Engine) RoutesDebug() {
	rs := en.Routes()
	for _, r := range rs {
		log.Println(r)
	}
}

func (en *Engine) Run() {
	log.Println("listening :8888")
	http.ListenAndServe(":8888", en)
}

func redirectTrailingSlash(c *Context) {
	req := c.R
	p := req.URL.Path
	if prefix := path.Clean(c.R.Header.Get("X-Forwarded-Prefix")); prefix != "." {
		prefix = regSafePrefix.ReplaceAllString(prefix, "")
		prefix = regRemoveRepeatedChar.ReplaceAllString(prefix, "/")

		p = prefix + "/" + req.URL.Path
	}
	req.URL.Path = p + "/"
	if length := len(p); length > 1 && p[length-1] == '/' {
		req.URL.Path = p[:length-1]
	}
	redirectRequest(c)
}

func redirectFixedPath(c *Context, root *node, trailingSlash bool) bool {
	req := c.R
	rPath := req.URL.Path

	if fixedPath, ok := root.findCaseInsensitivePath(rPath, trailingSlash); ok {
		req.URL.Path = BytesToString(fixedPath)
		redirectRequest(c)
		return true
	}
	return false
}

func redirectRequest(c *Context) {
	req := c.R
	rPath := req.URL.Path
	rURL := req.URL.String()

	code := http.StatusMovedPermanently // Permanent redirect, request with GET method
	if req.Method != http.MethodGet {
		code = http.StatusTemporaryRedirect
	}
	log.Printf("redirecting request %d: %s --> %s", code, rPath, rURL)
	c.Redirect(code, rURL)
}
