package main

import (
	"log"
	"net/http"
	"time"

	"github.com/cainmusic/rough/dev/v04"
)

func main() {
	r := rough.New()

	r.Use(func(c *rough.Context) {
		st := time.Now()
		c.Set("tnow", st)
		log.Println("start time", st.UnixNano())
		c.Next()
		et := time.Now()
		log.Println("end time", et.UnixNano())
		log.Println("use time", et.UnixMicro() - st.UnixMicro(), "microsecond")
	})

	r.Use(func(c *rough.Context) {
		log.Println("middleware without calling next")
	})

	r.Use(func(c *rough.Context) {
		t, err := c.Get("tnow")
		if err != nil {
			panic("ERROR TODO : " + err.Error())
		}
		//c.String(
		//	http.StatusOK,
		log.Printf(
			"hello, rough serves you. cur url is %s, method is %s, request time is %s",
			c.R.URL.Path,
			c.R.Method,
			t.(time.Time).String(),
		)
	})

	r.GET("/", func(c *rough.Context) {
		c.String(http.StatusOK, "use router")
	})

	r.Any("/any", func(c *rough.Context) {
		c.String(http.StatusOK, "you can use any method to request this url")
	})

	g1 := r.Group("/g1", func(c *rough.Context) {
		log.Println("this is a first middleware for group g1 with next called")
		c.Next()
		log.Println("end of the first middleware")
	}, func(c *rough.Context) {
		log.Println("this is a second middleware for group g1")
	}, func(c *rough.Context) {
		log.Println("this is a third middleware for group g1")
	})
	g1.GET("/abc", func(c *rough.Context) {
		c.String(http.StatusOK, "response from handler for /g1/abc")
	})

	r.Run()
}
