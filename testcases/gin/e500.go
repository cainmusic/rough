package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	//r := gin.New()
	r.GET("/", func(c *gin.Context) {
		a, b := 1, 0
		_ = a / b
		c.String(200, "123")
	})
	r.Run()
}
