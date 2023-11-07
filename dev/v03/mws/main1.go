package main

import (
	"log"
	"time"
)

type context struct {
	m map[string]any
}

func mw1(next func(context)) func(context) {
	return func(c context) {
		start := time.Now()
		log.Println("start time", start)
		next(c)
		end := time.Now()
		log.Println("end time", end)
		log.Println("request_time:", end.UnixNano()-start.UnixNano(), "ns")
	}
}

func mw2(next func(context)) func(context) {
	return func(c context) {
		log.Println("mw2 start")
		next(c)
		log.Println("mw2 end")
	}
}

func handler(c context) {
	log.Println("hello world")
}

func main() {
	c := context{}
	mw2(mw1(handler))(c)
}
