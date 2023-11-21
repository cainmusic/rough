package rough

import (
	"errors"
	"log"
	"math"
	"net/http"
	"net/url"

	"github.com/cainmusic/rough/dev/v07/render"
)

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
