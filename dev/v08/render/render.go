package render

import (
	"net/http"
)

type Render interface {
	Render(http.ResponseWriter) error
	WriteContentType(http.ResponseWriter)
}

var (
	_ Render = String{}
	_ Render = JSON{}
	_ Render = HTML{}
)

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}
