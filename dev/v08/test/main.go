package main

import (
	"net/http"

	"github.com/cainmusic/rough/dev/v08"
)

func main() {
	r := rough.New()

	r.Static("public", "./template")

	g1 := r.Group("g1")

	g1.Static("public", "./template")

	g1.Static("publi2", "./template")
	// 可以放在g1.Static("public", "./template")前面，放后面不生效
	//g1.Static("public2", "./template")

	r.Run()
}

func quickUrl(c *rough.Context) {
	c.String(http.StatusOK, "you just visited "+c.R.URL.String())
}
