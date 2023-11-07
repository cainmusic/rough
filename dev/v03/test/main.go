package main

import (
	"log"
	"net/http"
	"time"

	"github.com/cainmusic/rough/dev/v03"
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
	})

	r.Use(func(c *rough.Context) {
		log.Println("middleware without calling next")
	})

	r.Use(func(c *rough.Context) {
		t, err := c.Get("tnow")
		if err != nil {
			panic("ERROR TODO : " + err.Error())
		}
		c.String(
			http.StatusOK,
			"hello, rough serves you. cur url is %s, method is %s, request time is %s",
			c.R.URL.Path,
			c.R.Method,
			t.(time.Time).String(),
		)
	})

	r.Run()
}
