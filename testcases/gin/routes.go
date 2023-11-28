package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()

	r.Use(func(c *gin.Context) {
		log.Println("before1")
	})
	r.Use(func(c *gin.Context) {
		log.Println("before2")
	})

	g1 := r.Group("/g1")
	g1.GET("hello", func(c *gin.Context) { log.Println(111) }, quickResponseUrlString)

	r.Use(func(c *gin.Context) {
		log.Println("after1")
	})

	//r.GET("/index", quickResponseUrlString)

	r.Run()
}

func quickResponseUrlString(c *gin.Context) {
	c.String(http.StatusOK, "hello, gin serves you. you just visited %s", c.Request.URL.String())
}
