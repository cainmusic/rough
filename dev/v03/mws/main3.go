package main

import (
	"log"
)

type engine struct {
	handlers []func(*context)
}

func (e *engine) use(h func(*context)) {
	e.handlers = append(e.handlers, h)
}

func (e *engine) handle(c *context) {
	c.next()
}

type context struct {
	index    int
	handlers []func(*context)
}

func (c *context) next() {
	c.index++
	if c.index < len(c.handlers) {
		c.handlers[c.index](c)
	}
}

func ms1(c *context) {
	log.Println("ms1 start")
	c.next()
	log.Println("ms1 end")
}

func ms2(c *context) {
	log.Println("ms2 start")
	c.next()
	log.Println("ms2 end")
}

func handler(c *context) {
	log.Println("handler here")
}

func main() {
	e := &engine{
		handlers: []func(*context){},
	}
	e.use(ms1)
	e.use(ms2)

	e.use(handler)

	c := &context{
		index:    -1,
		handlers: e.handlers,
	}
	e.handle(c)
}
