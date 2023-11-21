package render

import (
	"fmt"
	"net/http"
)

type String struct {
	Format string
	Data   []any
}

var plainContentType = []string{"text/plain; charset=utf-8"}

func (r String) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	if len(r.Data) > 0 {
		_, err := fmt.Fprintf(w, r.Format, r.Data...)
		return err
	}
	_, err := w.Write([]byte(r.Format))
	return err
}

func (r String) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, plainContentType)
}
