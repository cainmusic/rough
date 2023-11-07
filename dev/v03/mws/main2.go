package main

import (
	"log"
)

type context int

var handler func(context) = func(c context) {
	log.Println("handler", c)
}

func setup(mwf func(c context, next func())) {
	tmp := handler
	handler = func(c context) {
		mwf(c, func() {
			tmp(c)
		})
	}
}

func mw1(c context, next func()) {
	log.Println(111)
	next()
	log.Println(222)
}

func mw2(c context, next func()) {
	log.Println(333)
	next()
	log.Println(444)
}

func main() {
	c := context(1)
	setup(mw1)
	setup(mw2)
	handler(c)
}
