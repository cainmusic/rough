package rough

import (
	"errors"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"path"
	"regexp"

	"github.com/cainmusic/rough/dev/v06/render"
)

/*
	others
*/

type HandleFunc func(*Context)

const EnKey = "__rough_engine"

const defaultMultipartMemory = 32 << 20

func GetEnKey(key string) string {
	return EnKey + "." + key
}

/*
	utils
*/

func lastChar(str string) uint8 {
	if str == "" {
		panic("The length of the string can't be 0")
	}
	return str[len(str)-1]
}

func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	if lastChar(relativePath) == '/' && lastChar(finalPath) != '/' {
		return finalPath + "/"
	}
	return finalPath
}

/*
	Context
*/

const abortIndex int8 = math.MaxInt8 >> 1

type Context struct {
	engine *Engine

	W    http.ResponseWriter
	R    *http.Request
	Keys map[string]any

	index    int8
	handlers []HandleFunc

	queryCache url.Values
	formCache  url.Values

	statusCode int
}

func (c *Context) Query(key string) string {
	value, _ := c.GetQuery(key)
	return value
}

func (c *Context) GetQuery(key string) (string, bool) {
	if values, ok := c.GetQueryArray(key); ok {
		return values[0], ok
	}
	return "", false
}

func (c *Context) QueryArray(key string) []string {
	values, _ := c.GetQueryArray(key)
	return values
}

func (c *Context) initQueryCache() {
	if c.queryCache == nil {
		if c.R != nil {
			c.queryCache = c.R.URL.Query()
		} else {
			c.queryCache = url.Values{}
		}
	}
}

func (c *Context) GetQueryArray(key string) ([]string, bool) {
	c.initQueryCache()
	values, ok := c.queryCache[key]
	return values, ok
}

func (c *Context) PostForm(key string) string {
	value, _ := c.GetPostForm(key)
	return value
}

func (c *Context) GetPostForm(key string) (string, bool) {
	if values, ok := c.GetPostFormArray(key); ok {
		return values[0], ok
	}
	return "", false
}

func (c *Context) PostFormArray(key string) []string {
	values, _ := c.GetPostFormArray(key)
	return values
}

func (c *Context) initFormCache() {
	if c.formCache == nil {
		c.formCache = make(url.Values)
		if err := c.R.ParseMultipartForm(defaultMultipartMemory); err != nil {
			// 无视"request Content-Type isn't multipart/form-data"的报错
			if !errors.Is(err, http.ErrNotMultipart) {
				log.Println("form parse error", err)
			}
		}
		c.formCache = c.R.PostForm
	}
}

func (c *Context) GetPostFormArray(key string) ([]string, bool) {
	c.initFormCache()
	values, ok := c.formCache[key]
	return values, ok
}

func (c *Context) Status(code int) {
	if code > 0 {
		c.W.WriteHeader(code)
	}
}

func (c *Context) String(code int, format string, values ...any) {
	c.Render(code, render.String{format, values})
}

func (c *Context) JSON(code int, obj any) {
	c.Render(code, render.JSON{Data: obj})
}

func (c *Context) HTML(code int, name string, obj any) {
	instance := c.engine.HTMLRender.Instance(name, obj)
	c.Render(code, instance)
}

func (c *Context) Redirect(code int, location string) {
	c.Render(-1, render.Redirect{
		Code:     code,
		Location: location,
		Request:  c.R,
	})
}

func (c *Context) Render(code int, r render.Render) {
	// 暂未考虑并发
	if c.statusCode != 0 {
		// TODO 处理警告
		log.Println("[warn] render already, skip")
		return
	}
	c.statusCode = code
	c.Status(code)
	if err := r.Render(c.W); err != nil {
		// TODO handle error
		//_ = c.Error(err)
		log.Println(err)
	}
	c.Abort()
}

func (c *Context) Set(key string, value any) {
	if c.Keys == nil {
		c.Keys = make(map[string]any)
	}
	c.Keys[key] = value
}

func (c *Context) Get(key string) (any, error) {
	v, ok := c.Keys[key]
	if ok {
		return v, nil
	}
	return nil, errors.New("key not found in context keys : " + key)
}

func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}

func (c *Context) Abort() {
	c.index = abortIndex
}

/*
	RouterGroup
*/

var (
	regEnLetter = regexp.MustCompile("^[A-Z]+$")

	anyMethods = []string{
		http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
		http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodConnect,
		http.MethodTrace,
	}
)

type RouterGroup struct {
	Handlers []HandleFunc
	basePath string
	engine   *Engine
}

func (group *RouterGroup) Group(relativePath string, handlers ...HandleFunc) *RouterGroup {
	return &RouterGroup{
		Handlers: group.combineHandlers(handlers),
		basePath: group.calculateAbsolutePath(relativePath),
		engine:   group.engine,
	}
}

func (group *RouterGroup) combineHandlers(handlers []HandleFunc) []HandleFunc {
	finalSize := len(group.Handlers) + len(handlers)
	mergedHandlers := make([]HandleFunc, finalSize)
	copy(mergedHandlers, group.Handlers)
	copy(mergedHandlers[len(group.Handlers):], handlers)
	return mergedHandlers
}

func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(group.basePath, relativePath)
}

func (group *RouterGroup) Use(middleware ...HandleFunc) {
	group.Handlers = append(group.Handlers, middleware...)
}

func (group *RouterGroup) handle(httpMethod, relativePath string, handlers []HandleFunc) {
	absolutePath := group.calculateAbsolutePath(relativePath)
	handlers = group.combineHandlers(handlers)
	group.engine.addRoute(httpMethod, absolutePath, handlers)
}

func (group *RouterGroup) Handle(httpMethod, relativePath string, handlers ...HandleFunc) {
	if matched := regEnLetter.MatchString(httpMethod); !matched {
		panic("http method " + httpMethod + " is not valid")
	}
	group.handle(httpMethod, relativePath, handlers)
}

// POST is a shortcut for router.Handle("POST", path, handlers).
func (group *RouterGroup) POST(relativePath string, handlers ...HandleFunc) {
	group.handle(http.MethodPost, relativePath, handlers)
}

// GET is a shortcut for router.Handle("GET", path, handlers).
func (group *RouterGroup) GET(relativePath string, handlers ...HandleFunc) {
	group.handle(http.MethodGet, relativePath, handlers)
}

// DELETE is a shortcut for router.Handle("DELETE", path, handlers).
func (group *RouterGroup) DELETE(relativePath string, handlers ...HandleFunc) {
	group.handle(http.MethodDelete, relativePath, handlers)
}

// PATCH is a shortcut for router.Handle("PATCH", path, handlers).
func (group *RouterGroup) PATCH(relativePath string, handlers ...HandleFunc) {
	group.handle(http.MethodPatch, relativePath, handlers)
}

// PUT is a shortcut for router.Handle("PUT", path, handlers).
func (group *RouterGroup) PUT(relativePath string, handlers ...HandleFunc) {
	group.handle(http.MethodPut, relativePath, handlers)
}

// OPTIONS is a shortcut for router.Handle("OPTIONS", path, handlers).
func (group *RouterGroup) OPTIONS(relativePath string, handlers ...HandleFunc) {
	group.handle(http.MethodOptions, relativePath, handlers)
}

// HEAD is a shortcut for router.Handle("HEAD", path, handlers).
func (group *RouterGroup) HEAD(relativePath string, handlers ...HandleFunc) {
	group.handle(http.MethodHead, relativePath, handlers)
}

// Any registers a route that matches all the HTTP methods.
// GET, POST, PUT, PATCH, HEAD, OPTIONS, DELETE, CONNECT, TRACE.
func (group *RouterGroup) Any(relativePath string, handlers ...HandleFunc) {
	for _, method := range anyMethods {
		group.handle(method, relativePath, handlers)
	}
}

/*
	H
*/

type H map[string]any

/*
	Engine
*/

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
