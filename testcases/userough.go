package main

import (
	"log"
	"net/http"

	"github.com/cainmusic/rough"
)

func main() {
	r := rough.New()

	r.Static("/gin_cases", "./gin")

	r.GET("/", func(c *rough.Context) {
		c.String(http.StatusOK, "hello, rough serves you.")
	})

	r.GET("/hello", quickResponseUrlString)

	r.GET("/query", func(c *rough.Context) {
		log.Println(c.Query("a"))
	}, quickResponseUrlString)

	r.POST("/postform", func(c *rough.Context) {
		log.Println(c.PostForm("b"))
	}, quickResponseUrlString)

	r.GET("/hello/:x/*y", func(c *rough.Context) {
		log.Println(c.Param("x"))
		log.Println(c.Param("y"))
	}, quickResponseUrlString)

	r.RoutesDebug()

	r.Run() // :8888
}

func quickResponseUrlString(c *rough.Context) {
	c.String(http.StatusOK, "you just visited "+c.R.URL.String())
}
