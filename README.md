# rough
rough version gin http server

# 介绍

项目目录下`rough.go`代码一般使用`./dev`目录下最高版本目录下的`rough.go`代码

项目详情请看`doc.md`

# 使用rough

`./testcases/userough.go`

``` go
package main

import "github.com/cainmusic/rough"

func main() {
	r := rough.New()
	r.GET("/", func(c *rough.Context) {
		c.String(200, "hello, rough serves you.")
	})
	r.Run() // :8888
}

```

访问`http://localhost:8888/`：

```
hello, rough serves you.
```

# 实现

* v1，初始化和启动
* v2，上下文
* v3，中间件
* v4，路由