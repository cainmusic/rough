package rough

import (
	"net/http"
)

type Engine struct {
}

func New() *Engine {
	return &Engine{}
}

func (en *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello, rough serves you."))
}

func (en *Engine) Run() {
	http.ListenAndServe(":8888", en)
}
