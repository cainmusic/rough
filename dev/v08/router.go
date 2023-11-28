package rough

import (
	"net/http"
	"regexp"
	"strings"
)

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

func (group *RouterGroup) Static(relativePath, root string) {
	absolutePath := group.calculateAbsolutePath(relativePath)
	fileServer := http.StripPrefix(absolutePath, http.FileServer(http.Dir(root)))
	group.engine.Use(func(c *Context) {
		if (c.R.Method == "GET" || c.R.Method == "HEAD") &&
			strings.HasPrefix(c.R.URL.Path, absolutePath) {
			fileServer.ServeHTTP(c.W, c.R)
			c.Abort()
			return
		}
		c.Next()
	})
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
