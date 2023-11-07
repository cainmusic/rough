package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		log.Println("before1")
	})
	r.Use(func(c *gin.Context) {
		log.Println("before2")
	})
	r.GET("/", func(c *gin.Context) {
		c.String(200, "hello /")
	})
	r.Use(func(c *gin.Context) {
		log.Println("after1")
	})
	r.GET("/123", func(c *gin.Context) {
		c.String(200, "123")
	})
	r.Run()
}
