package main

import (
	"log"
	"net/http"

	"github.com/cainmusic/rough/dev/v09"
)

func main() {
	r := rough.New()

	r.GET("/", quickUrl)
	r.GET("/vendor/:x/*y", func(c *rough.Context) {
		log.Println(c.Param("x"))
		log.Println(c.Param("y"))
	}, quickUrl)

	r.Static("/static", "./static")

	r.GET("/hello/u:id", func(c *rough.Context) {
		log.Println(c.Param("id"))
	}, quickUrl)

	r.RoutesDebug()

	r.Run()
}

func quickUrl(c *rough.Context) {
	c.String(http.StatusOK, "you just visited "+c.R.URL.String())
}
