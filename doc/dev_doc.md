# 【一】初衷

很多人在学习了编程语言之后会陷入一个迷茫期，会遇到一个问题：我学了这个语言可以干啥？

很多人会有基本的写代码的能力但缺乏组织项目的能力，可以对框架一知半解做一个CRUD BOY，但基本说不清框架的原理，更不提实现。

本文的主要目的就是从基础编码到组织项目，从解析框架到实现框架，来学习一下一个框架是如何诞生的。

# 【二】Gin简介

本次项目主要目标是实现一个粗糙的类Gin的http框架，这里会对Gin做一个简单的介绍。

（注：简单的了解可能不足以支撑后面的阅读，还是建议看下Gin的文档并实现一些简单功能。）

## 【二。一】最简单的Gin服务

``` go
// ./testcases/hello.go
package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(200, "hello world")
	})

	r.Run()
}
```

`go run hello.go`启动之后，在浏览器访问`http://localhost:8080/`，之后可以在浏览器看到文字`hello world`，并可以在命令行看到日志：

```
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

[GIN-debug] GET    /                         --> main.main.func1 (3 handlers)
[GIN-debug] [WARNING] You trusted all proxies, this is NOT safe. We recommend you to set a value.
Please check https://pkg.go.dev/github.com/gin-gonic/gin#readme-don-t-trust-all-proxies for details.
[GIN-debug] Environment variable PORT is undefined. Using port :8080 by default
[GIN-debug] Listening and serving HTTP on :8080
[GIN] 2023/10/16 - 15:02:53 | 200 |      42.416µs |             ::1 | GET      "/"
[GIN] 2023/10/16 - 15:02:53 | 404 |       1.125µs |             ::1 | GET      "/favicon.ico"
```

## 【二。二】拆解

上面的日志我们只关注：

```
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] GET    /                         --> main.main.func1 (3 handlers)

[GIN-debug] Environment variable PORT is undefined. Using port :8080 by default
[GIN-debug] Listening and serving HTTP on :8080
```

### 【二。二。一】初始化

第一段日志提到`Creating an Engine instance with the Logger and Recovery middleware already attached`，这段日志对应代码的`gin.Default()`，是服务实例的初始化。

`gin.Default()`代码创建了一个`*gin.Engine`的实例，这个创建过程实际调用的是`New`函数，下面是源码：

``` go
func New() *Engine {
	debugPrintWARNINGNew()
	engine := &Engine{
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},
		FuncMap:                template.FuncMap{},
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      false,
		HandleMethodNotAllowed: false,
		ForwardedByClientIP:    true,
		RemoteIPHeaders:        []string{"X-Forwarded-For", "X-Real-IP"},
		TrustedPlatform:        defaultPlatform,
		UseRawPath:             false,
		RemoveExtraSlash:       false,
		UnescapePathValues:     true,
		MaxMultipartMemory:     defaultMultipartMemory,
		trees:                  make(methodTrees, 0, 9),
		delims:                 render.Delims{Left: "{{", Right: "}}"},
		secureJSONPrefix:       "while(1);",
		trustedProxies:         []string{"0.0.0.0/0", "::/0"},
		trustedCIDRs:           defaultTrustedCIDRs,
	}
	engine.RouterGroup.engine = engine
	engine.pool.New = func() any {
		return engine.allocateContext(engine.maxParams)
	}
	return engine
}

func Default() *Engine {
	debugPrintWARNINGDefault()
	engine := New()
	engine.Use(Logger(), Recovery())
	return engine
}
```

`New`函数简单的创建了`engine`对象，`Default`函数获得这个`engine`对象后，`Use`了两个中间件`Logger()`和`Recovery()`，之后就返回了。

在此我们不详细解释这些内容，但让我们先记住两个概念：`gin.Engine`和中间件。

### 【二。二。二】路由

第二段日志`GET    /                         --> main.main.func1 (3 handlers)`对应的是代码中的`r.GET("/", func(c *ginContext){...})`，这是我们定义的一条路由信息。

所谓路由，简单理解就是“路径信息”，指示从哪到哪的信息。

`r.GET("/", func(c *ginContext){...})`如何理解？

1. `GET`是请求方法，可以类比“步行”、“公交”
2. `"/"`是目的地
3. 后面的`func`是处理过程，大概就是在`GET`的方法下，到达`"/"`需要进行哪些处理

实践中我们可以定义更多的“路径信息”，所有这些路由组成了复杂的服务。

### 【二。二。三】启动

第三段日志`Using port :8080 by default`和`Listening and serving HTTP on :8080`对应代码中的`r.Run()`。

`Run`函数首先处理了地址信息，之后启动了`http`的服务，核心代码是：

``` go
http.ListenAndServe(address, engine.Handler())
```

而这个`http`来自哪里？来自标准库`net/http`。

`Gin`框架实际使用的是`net/http`标准库做底层服务的。

## 【二。三】Gin小结

根据前面的简单案例的拆解，我们可以大致总结几点：

1. Gin框架使用标准库`net/http`做底层服务
2. Gin框架实现了路由、中间件等功能
3. Gin框架的服务初始化、启动很方便（凑数）

# 【三】标准库net/http

我们要写一个类`Gin`框架，而`Gin`框架又是用的`net/http`标准库，那我们起码要了解一点`net/http`怎么用。

## 【三。一】ListenAndServe

首先看上面的`http.ListenAndServe()`函数，下面是源码中相关的一些内容：

``` go
// ListenAndServe函数签名
func ListenAndServe(addr string, handler Handler) error

// Handler接口
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}

// ResponseWriter接口
type ResponseWriter interface {...}

// Request结构体
type Request struct {...}

// HandlerFunc
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
    f(w, r)
}

// ServeMux
type ServeMux struct {...}

func (mux *ServeMux) ServeHTTP(w ResponseWriter, r *Request)
```

`net/http`实际提供了一系列处理`http`请求的方法但`Gin`框架用到的内容不多。

`net/http`有自己的路由结构`type ServeMux`，`Gin`则实现了自己的`type RouterGroup`。

我们这里主要看`http.ListenAndServe()`相关内容。

`ListenAndServe`的函数签名是`func ListenAndServe(addr string, handler Handler) error`。

我们可以看到这个函数接收一个地址和一个`handler`，返回一个`error`。

返回`error`不提，地址一般是一个机器（IP、域名、hostname等）+端口，`handler`是一个`Handler`类型，而`Handler`类型是一个接口`type Handler interface`，这个接口声明了一个`ServeHTTP(ResponseWriter, *Request)`方法。

在`net/http`标准库创建的服务中，服务接收到一个新的请求时，经过一系列代码调用，最终调用的就是上面`handler`这个`Handler`接口的`ServeHTTP`函数。

## 【三。二】Handler和ServeHTTP

由`net/http`包创建的服务处理`http`请求主要用到的就是`Handler`接口的`ServeHTTP`方法。

让我们来看下面的代码：

``` go
// ./testcases/multi.go
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

```

对于`localhost:8001`有：

```
访问http://localhost:8001/
404 page not found

访问http://localhost:8001/ping
pong
```

对于`localhost:8002`有：

```
访问http://localhost:8002/
you just visited /

访问http://localhost:8002/123
you just visited /123
```

对于`localhost:8003`有：

```
访问http://localhost:8003/
hello, gin serves you.

访问http://localhost:8003/123
404 page not found
```

上面三个例子就是`net/http`包的三种基本用法，接下来详细介绍。

### 【三。二。一】http.ServeMux

方法一使用的实际上是`net/http`本身提供的`ServeMux`结构

调用`http.Handle`函数实际上调用的是`DefaultServeMux.Handle()`。

调用`http.HandleFunc`函数实际上调用的是`DefaultServeMux.HandleFunc()`。

`HandleFunc`方法最终也调用`Handle`：

``` go
func (mux *ServeMux) HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
    ...
    mux.Handle(pattern, HandlerFunc(handler))
}
```

这里我们可以一窥`HandlerFunc`这个类型的用法，不过暂时不细说。

我们可以看到使用`net/http`包的第一个方法就是使用`http.ServeMux`结构，本例中我们使用的是默认的`ServeMux`，在实际使用的时候我们也可以自己组织这个结构，使用`NewServeMux()`函数可以获得一个该结构供你自己组织。

### 【三。二。二】http.HandlerFunc

上例我们提了一嘴`HandlerFunc`这个类型。

和这个类型相关的代码在上面已经列出来过，这里再次列出如下：

``` go
// HandlerFunc
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
    f(w, r)
}
```

接下来的内容会涉及Go语言的类型、接口、类型转换等内容。

首先，`net/http`包中定义了一个函数签名类型`HandlerFunc`，其签名是`func(ResponseWriter, *Request)`。并且这个类型实现了一个方法`ServeHTTP`。

前面我们知道`Handler`接口只要实现`ServeHTTP`就可以实现：

``` go
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}
```

于是我们知道，`HandlerFunc`这个类型，实现了`Handler`接口。

现在，我们有一个实现了签名`func(ResponseWriter, *Request)`的函数，无论这个函数叫什么名字，这个函数是符合`HandlerFunc`类型的定义的。

那么，情形如下：

* 我们有一个函数，实现了签名`func(ResponseWriter, *Request)`，符合`HandlerFunc`类型的定义
* `HandlerFunc`类型实现了`Handler`接口
* 我们想把这个函数当作`Handler`来用

于是有：

``` go
SomeFunc := func(ResponseWriter, *Request) {...}
var _ http.Handler = http.HandlerFunc(SomeFunc)
```

再来看`HandlerFunc`的`ServeHTTP`方法是调用这个类型函数本身。

是不是很神奇，我们把一个函数转化成了一个自我调用的接口，我们管`http.HandlerFunc`类型叫做适配器。

（注：如果你不想用`http.HandlerFunc`类型转换处理这个问题，你可以怎么做？你可以写一个空结构体或者其他什么类型，并实现一个`ServeHTTP`方法，没错，这就是后面的Gin）

### 【三。二。三】http.Handler

在聊Gin之前，我们插一嘴聊一下`Handler`接口。

这个接口可以说贯穿`net/http`包，前面提到的`http.ServeMux`和`http.HandlerFunc`都实现了这个接口，后面的`gin.Engine`也实现了这个接口。

而`net/http`包中可以使用`Handler`接口的地方也很多，其中一个就是方法二中提到的`ListenAndServe`函数，当然还有一个就是方法一中用到的`Handle`方法。

从使用的角度来说，这两个场景是不完全相同的，方法一中的`Handler`服务于一个`url`，而方法二中的`Handler`服务于一整个服务，但他们为何用到了相同的接口呢？

因为，从逻辑上来说，方法二的服务范围是包含方法一的，它们做的事情是类似的：

* 相同：方法二和方法一，从行为层面，都是服务于一个`request`，返回一个`response`
* 区别：方法二可以服务于传入的所有请求，方法一的那一个`Handler`服务于传入的某一类请求，你还可以定义新的`Handler`服务其他请求

所有接收请求，处理（或传递）响应的方法都是`Handler`，`Handler`之间可以有包含关系，这种包含传递的思想本身就是一种中间件思想，也有人管这叫做洋葱结构。

参考：

``` go
// net/http包，ServeMux类型的ServeHTTP方法
func (mux *ServeMux) ServeHTTP(w ResponseWriter, r *Request) {
	if r.RequestURI == "*" {
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(StatusBadRequest)
		return
	}
	h, _ := mux.Handler(r)
	h.ServeHTTP(w, r)
}
```

也参考：

``` go
// https://zhuanlan.zhihu.com/p/354482838
// 代码来自于上面的链接，没测试过，仅供参考
package main

import (
    "context"
    "net/http"
    "github.com/google/uuid"
    "log"
)

func main() {
    http.Handle("/", inejctMsgID(http.HandlerFunc(HelloWorld)))
    log.Fatalln(http.ListenAndServe(":8080", nil))
}

func HelloWorld(w http.ResponseWriter, r *http.Request) {
    msgID :=""
    if m :=r.Context().Value("msgId"); m != nil {
        if value, ok := m.(string); ok {
            msgID=value
        }
    }
    w.Header().Add("msgId", msgID)
    w.Write([]byte("Hello,世界"))
}

func inejctMsgID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        msgID :=uuid.New().String()
        ctx :=context.WithValue(r.Context(),"msgId", msgID)
        req := r.WithContext(ctx)
        next.ServeHTTP(w,req)
    })
}
```

关于中间件的实现，后面会详细说。不过中间件的实现方法有很多，后续会按需求介绍。

### 【三。二。四】gin.Engine

`gin.Engine`同上面两个实现都不同，但同第二个更接近。

与第一个方法使用默认的`ServeMux`不同，后两个方法都是直接使用`ListenAndServe`函数传入了`Handler`。

区别在于方法二中，传入的是`http.Handler`类型转换的函数，而本方法中传的是Gin框架中的实现了`Handler`接口的`*gin.Engine`对象。

路由、中间件等其他服务基本全部由Gin框架自己实现。

# 【四】Rough

我们创建一个包，叫Rough，来逐步实现一个利用`net/http`库来实现`http`服务的框架。

为什么叫Rough，包如其名：粗糙的、大致的、草稿。Rough是一个粗糙的Gin。

## 【四。一】初始化和启动

代码：

```
$ tree ./dev/v01/
./dev/v01
├── rough.go
└── test
    └── main.go

1 directory, 2 files
```

启动服务，请求`http://localhost:8888/`，可以看到：

```
hello, rough serves you.
```

如此，你获得了一个最简单的rough框架，无论你请求什么接口，都会返回你上面的内容。

## 【四。二】上下文

`gin.Context`是一个请求上下文的集合，几乎涵盖了请求过程的全部信息。

`net/http`包使用`Handler`处理请求，调用`ServeHTTP`方法仅传递`ResponseWriter`和`*Request`两个对象，这里有两个问题：

1. 如果我们需要传递额外的资源，如何将资源传入处理环境
2. 要了解`ResponseWriter`和`*Request`提供的方法来进行处理，学习成本较高，且方法不一定易用

于是Gin框架提供了`gin.Context`来抽象整个请求周期的资源总和并提供一些易用方法。

定义：
``` go
type Context struct {
	W    http.ResponseWriter
	R    *http.Request
	Keys map[string]any
}
```

包含`net/http`包的`ResponseWriter`和`*Request`和一个自定义的`map`结构`Keys`。

对其中的`Keys`我们提供`Set(string, any)`和`Get(string) (any, error)`方法来存取请求过程中共享的内容。

而对`ResponseWriter`我们提供`Status(int)`和`String(int, string, ...any)`方法来简化响应的调用。

后续还会继续丰富`Context`的方法，新增方法将不再详细介绍。

之后我们改写`ServeHTTP`方法，生成一个当前请求周期使用的`Context`对象并传入处理函数。

我们将当前时间传入`Context`并在之后的请求中取出使用，这样就可以在请求过程中共享内容。

代码：

```
tree ./dev/v02/
./dev/v02/
├── rough.go
└── test
    └── main.go

1 directory, 2 files
```

启动服务，请求`http://localhost:8888/`，可以看到：

```
hello, rough serves you. cur url is /, method is GET, request time is 2023-10-26 17:28:41.201787 +0800 CST m=+19.141528585
```

## 【四。三】中间件

中间件的概念其实很丰富，我们这里主要讲洋葱型的中间件，大概的形式如下：

```
// 伪代码
func1 () {
	before1
	func2 () {
		before2
		func3 () {
			before3
			handler () {
				handler request
			}
			after3
		}
		after2
	}
	after1
}
```

构建这样一个结构的方法有很多，我们这里分别会在示例和`rough`中实现几种简单做法，`gin`框架实现的是另外的一种做法。

代码路径`./dev/v03/mws/main1.go`，复制如下：

``` go
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
```

代码路径`./dev/v03/mws/main2.go`，复制如下：

``` go
package main

import (
	"log"
)

type context int

var handler func(context) = func(c context) {
	log.Println("handler", c)
}

func setup(mwf func(context, func())) {
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

```

从上面的代码可以看出，一般`handler`函数都是一个`func(c1, c2, c3)`的形式，比如：

* 上面例子的`func(context)`
* `net/http`包的`func(ResponseWriter, *Request)`
* `gin`框架的`func(*Context)`
* `rough`仿`gin`的`func(*Context)`（`v02`例子中的`handleRequest`方法）

在中间件的不同的实现中，`handler`函数的构造都差不多，但调用方式有些许区别：

`main1.go`和`main2.go`提供了两种基本的通过中间件函数构建最终处理程序的方式，只不过两个方法对中间件的抽象程度不同。
`main1.go`的中间件函数通过传入一个被包围`handler`，生成一个新的`handler`，来构建最终处理程序。
`main2.go`则聚焦当前中间件的职能，获取上下文以及功能性`next`句柄来构建中间件操作，最后再用构建函数将中间件装入处理程序。

`gin`框架的中间件实现简单实现跟上面两种方法都不太一样，大概实现一下是下面的感觉。

代码路径`./dev/v03/mws/main3.go`，复制如下：

``` go
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
```

可以看到，中间件的核心操作是处理`next`。

不过，需要注意的是，把对`next`的调用作为处理中间件的核心，是我自己对中间件的总结。实际上`gin`框架对`handlers`调用链进行调用的实现中，除了在中间件中主动调用`next`之外，还有自动调用下一个组件的方法。

那是因为`gin.Context`的`Next`方法的实现如下：

``` go
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}
```

可以看到，一般情况下，如果所有`handler`都调用了`Next`，则上述函数和我们`main3.go`中的`next`方法有着相同的调用链。但如果某个`handler`没有调用`Next`，而`c.handlers`中还有未执行的`handler`，则会自动执行下一个`handler`，直到`c.handlers`中所有的`handler`都被执行（或者`Abort`被提前执行）。

那么，接下来让我们来实现一个类似`gin`框架的中间件调用方法。

虽然中间件调用和路由关联较大，但本节仅会解决中间件调用问题，路由功能会在后续章节中实现。

代码：

```
./dev/v03
├── mws
│   ├── main1.go
│   ├── main2.go
│   └── main3.go
├── rough.go
└── test
    └── main.go

2 directories, 5 files
```

启动服务，请求`http://localhost:8888/`，可以看到：

```
hello, rough serves you. cur url is /, method is GET, request time is 2023-10-27 17:11:49.970086 +0800 CST m=+9.297035709
```

在标准输出中可以看到：

```
2023/10/27 17:11:49 start time 1698397909986068000
2023/10/27 17:11:49 middleware without calling next
2023/10/27 17:11:49 end time 1698397909986125000
```

## 【四。四】路由

路由的核心是数据结构构建。

`gin`中最基本的路由设置大概长下面的样子：

``` go
	r.GET("/", func(c *gin.Context) {
		c.String(200, "hello world")
	})

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.GET("/echo/:str", func(c *gin.Context){
		c.String(200, c.Param("str"))
	})
```

上面三个路由分别定义了三个模式对应的处理方法。

当然了，考虑到可能存在的中间件，此路由的定义并不表示该模式仅由对应的函数处理。

由`GET`、`POST`、`DELETE`、`PUT`等方法定义的路由配置函数，大部分指定的是路由链中的最后节点，前面可能配置有若干中间件。

整个路由的核心大概就是配置前面的路由终节点和其前面的中间件的一整个数据结构了。

前面在【中间件】章节中我们实际上已经使用了一种最简单的路由数据结构：`handlers []HandleFunc`。

但这个结构仅仅是将所有中间件和最终的单一处理函数链接了起来，甚至没法处理最简单的`url`路由。

在比较简单的路由中，可以直接使用`map[string]HandleFunc`来存储路由方法。

但这个方法很难适应复杂的需求。

下面我们来学习一下`gin`的路由数据结构构建。

### 【四。四。一】trees

`trees`定义如下

``` go
type methodTree struct {
	method string
	root   *node
}

type methodTrees []methodTree
```

`method`就是`http`的几个方法：`GET`、`POST`、`DELETE`、`PUT`等。
`root`指向的是`tree`的根节点。

### 【四。四。二】tree

我们没有`tree`这样一个结构，我们有的是节点`node`，所谓的`tree`是由多个`node`构成一棵树：`radix tree`。

```
type node struct {
	path      string
	indices   string
	wildChild bool
	nType     nodeType
	priority  uint32
	children  []*node // child nodes, at most 1 :param style node at the end of the array
	handlers  HandlersChain
	fullPath  string
}
```

那么，这棵树怎么组织使用，代码在`gin`项目跟目录`tree.go`文件中的下面两个方法：

```
func (n *node) addRoute(path string, handlers HandlersChain)

func (n *node) getValue(path string, params *Params, skippedNodes *[]skippedNode, unescape bool) (value nodeValue)
```

以这两个方法为切入口，可以详细了解`gin`项目的`radix tree`是如何组织的。

单纯的`radix tree`本身是不复杂的，但因为`gin`项目还要在其中实现通配符`:`和`*`，所以代码会复杂一点。

可以参考`./dev/v04/radixtree/tree.go`文件中的注释。

### 【四。四。三】rough简单实现

在我们的`rough`项目中，并不会完整实现`gin`的全部路由功能，我们仅简单实现不含通配符的路由，数据结构用`map`实现。

虽然不实现`radix tree`结构，但会实现`RouterGroup`来辅助处理路由。

代码：

```
./dev/v04
├── radixtree
│   └── tree.go
├── rough.go
└── test
    └── main.go

2 directories, 3 files
```

启动服务，请求各个接口，可以看到：

```
http://localhost:8888/
use router

http://localhost:8888/any
you can use any method to request this url

http://localhost:8888/g1
Not Found

http://localhost:8888/g1/abc
response from handler for /g1/abc

http://localhost:8888/g1/abc/
Not Found
```

在命令行使用curl命令用不同的method进行请求：

```
$ curl -X POST "http://localhost:8888/any"
you can use any method to request this url

$ curl -X PATCH "http://localhost:8888/any"
you can use any method to request this url

$ curl -X DELETE "http://localhost:8888/any1"
Not Found
```

# 【五】添油加醋

目前`rough`尚不能满足日常使用，需要不断的增加内容。

## 【五。一】Query和PostForm

`gin`获取`query`是通过`net/http`包的`Request`的`URL`获取的，`URL`来自于`net/url`包，有`Query`方法。

`gin`获取`postform`是通过`net/http`包的`Request`的`ParseMultipartForm`方法获得，`ParseMultipartForm`调用`ParseForm`，`ParseMultipartForm`支持上传文件，但`rough`暂时不支持。

代码：

```
./dev/v05
├── rough.go
└── test
    └── main.go

1 directory, 2 files
```

启动服务，在命令行使用curl命令用不同的method进行请求：

```
$ curl http://localhost:8888/query?a=1
you just visited /query?a=1

$ curl -X POST -d 'b=111' "http://localhost:8888/postform"
you just visited /postform
```

日志打印：

```
2023/11/08 19:46:33 GET /query 2 handlers
2023/11/08 19:46:33 1
2023/11/08 19:47:16 POST /postform 2 handlers
2023/11/08 19:47:16 111
```

## 【五。二】JSON、HTML和Redirect

目前`rough`仅有`String`的响应方式，我们这里增加两个响应格式，以及一个重定向的功能。

`gin`的响应方法可以多次调用，可写的`content`会一直往响应信息里写，但`http`的状态码不能覆盖，会报警。

但`rough`为了避免使用上的歧义，会限制响应方法仅一次调用有效。

代码：

```
./dev/v06
├── render
│   ├── html.go
│   ├── json.go
│   ├── redirect.go
│   ├── render.go
│   └── text.go
├── rough.go
├── template
│   ├── files
│   │   ├── 1.html
│   │   └── 2.html
│   └── index.html
└── test
    └── main.go

4 directories, 10 files
```

启动服务，请求各个接口，可以看到：

```
http://localhost:8888/visit_index
这是index.html，一个html模板 你好

http://localhost:8888/some_json
{"age":18,"name":"小王"}

http://localhost:8888/some_json_2
{"name":"小张","age":17}

http://localhost:8888/redirect
在浏览器访问会直接跳转http://localhost:8888/visit_index

$ curl http://localhost:8888/redirect
<a href="/visit_index">Temporary Redirect</a>.
```

## 【五。三】拆分`rough.go`

目前来看`rough.go`并不算大，但考虑到使代码更容易阅读，拆出三个文件：

```
context.go 上下文
router.go 路由
utils.go 工具
```

代码：

```
./dev/v07
├── context.go
├── render
│   ├── html.go
│   ├── json.go
│   ├── redirect.go
│   ├── render.go
│   └── text.go
├── rough.go
├── router.go
├── template
│   ├── files
│   │   ├── 1.html
│   │   └── 2.html
│   └── index.html
├── test
│   └── main.go
└── utils.go

4 directories, 13 files
```

## 【五。四】静态文件

`gin`的静态文件使用路由通配符实现的，但`rough`还没实现通配符，所以不应用路由的方法实现静态文件。

（注意是不应不是无法，非要用路由方法，我们可以给每个文件加一个路由，但这样非常不好）

目前我们可以考虑用中间件匹配`url`来实现。

代码：

```
./dev/v08
├── context.go
├── render
│   ├── html.go
│   ├── json.go
│   ├── redirect.go
│   ├── render.go
│   └── text.go
├── rough.go
├── router.go
├── template
│   ├── files
│   │   ├── 1.html
│   │   └── 2.html
│   └── index.html
├── test
│   └── main.go
└── utils.go

4 directories, 13 files
```
