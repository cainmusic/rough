package main

import (
	"github.com/gin-gonic/gin"

	"log"
	"net/http"
)

func main() {
	r := gin.Default()

	//r.GET("/:function", quickResponseUrlString)

	// param不排斥兄弟节点
	//r.GET("/user/biliabc", quickResponseUrlString)
	r.GET("/user/biliabc/", quickResponseUrlString)

	r.GET("/user/bili:id", func(c *gin.Context) {
		log.Println(c.Param("id"))
	}, quickResponseUrlString)

	//r.GET("/user/bili:ids/", quickResponseUrlString)
	r.GET("/user/bili:id/", quickResponseUrlString)

	// 跟:id冲突
	//r.GET("/user/bili:name", quickResponseUrlString)

	// 不冲突
	r.GET("/user/bili:id/hello", quickResponseUrlString)
	// 冲突
	//r.GET("/user/bili:name/hello", quickResponseUrlString)

	// 两个/之间不能有多个通配符
	//r.GET("/book/bili:id:name", quickResponseUrlString)

	// catchAll的前面必须是/
	//r.GET("/static/hello*file", quickResponseUrlString)

	// catchAll排斥兄弟节点
	//r.GET("/static/aaa:file", quickResponseUrlString)
	//r.GET("/static/hello", quickResponseUrlString)

	// 会导致index out of range的报错
	//r.GET("/static/", quickResponseUrlString)
	r.GET("/static/*file", func(c *gin.Context) {
		log.Println(c.Param("file"))
	}, quickResponseUrlString)

	// 都报错，catchAll后面不能有任何额外信息
	//r.GET("/html/*html/hello", quickResponseUrlString)
	//r.GET("/html/*html/", quickResponseUrlString)

	go server2()

	go server3()

	r.Run()
}

func server2() {
	r := gin.Default()

	r.GET("/*function", quickResponseUrlString)

	r.Run(":8081")
}

func server3() {
	r := gin.Default()
	// panic: runtime error: index out of range [0] with length 0
	//r.GET("/static/", func(c *gin.Context) { c.String(200, "static") })
	r.GET("/static", func(c *gin.Context) { c.String(200, "static") })
	r.GET("/static/*file", func(c *gin.Context) { c.String(200, "static file") })

	r.Run(":8082")
}

func quickResponseUrlString(c *gin.Context) {
	c.String(http.StatusOK, "hello, gin serves you. you just visited %s", c.Request.URL.String())
}
