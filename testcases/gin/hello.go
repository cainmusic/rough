package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(200, "hello world")
	})

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.GET("/echo/:str", func(c *gin.Context) {
		c.String(200, c.Param("str"))
	})

	r.GET("/call_render_several_times", func(c *gin.Context) {
		c.String(200, "hello")
		c.String(200, "abc")
		c.JSON(200, gin.H{"a": 1, "b": 2})
		c.String(100, "def")
		c.String(500, "ghi")
	})

	r.Run()
}
