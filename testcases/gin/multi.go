package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	go defaultHttpServeMux()
	go handlerFunc()
	go ginServer()
	for {
	}
}

func defaultHttpServeMux() {
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})
	http.ListenAndServe(":8001", nil)
}

func handlerFunc() {
	http.ListenAndServe(":8002", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("you just visited " + r.URL.String()))
	}))
}

func ginServer() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(200, "hello, gin serves you.")
	})

	r.Run(":8003")
}
