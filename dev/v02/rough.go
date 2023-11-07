package rough

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Engine struct {
}

type Context struct {
	W    http.ResponseWriter
	R    *http.Request
	Keys map[string]any
}

func (c *Context) Status(code int) {
	c.W.WriteHeader(code)
}

func (c *Context) String(code int, format string, values ...any) {
	c.Status(code)
	c.W.Write([]byte(fmt.Sprintf(format, values...)))
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

func New() *Engine {
	return &Engine{}
}

const EnKey = "__rough_engine"

func GetEnKey(key string) string {
	return EnKey + "." + key
}

func (en *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := &Context{
		W: w,
		R: r,
	}
	c.Set(GetEnKey("tnow"), time.Now())
	en.handleRequest(c)
}

func (en *Engine) handleRequest(c *Context) {
	t, err := c.Get(GetEnKey("tnow"))
	if err != nil {
		panic("ERROR TODO : " + err.Error())
	}
	c.String(
		http.StatusOK,
		"hello, rough serves you. cur url is %s, method is %s, request time is %s",
		c.R.URL.Path,
		c.R.Method,
		t.(time.Time).String(),
	)
}

func (en *Engine) Run() {
	http.ListenAndServe(":8888", en)
}
