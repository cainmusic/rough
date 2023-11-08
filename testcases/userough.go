package main

import "github.com/cainmusic/rough"

func main() {
	r := rough.New()

	r.GET("/", func(c *rough.Context) {
		c.String(200, "hello, rough serves you.")
	})

	r.GET("/hello", func(c *rough.Context) {
		c.String(200, c.R.URL.String())
	})

	r.Run() // :8888
}
