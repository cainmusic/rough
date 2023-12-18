使用`go get -u github.com/cainmusic/rough`获取最新代码

20231108

* `ServeHTTP`时使用`URL.Path`（而不是`URL.String()`）获取路由

20231108

* 增加了对`Query`和`PostForm`的解析

20231121

* 增加了`JSON`、`HTML`、`Redirect`三种响应方法

* 拆分`rough.go`文件，没有功能变化

20231128

* 增加了对静态文件的响应支持

20231218

* 使用了支持通配符的radix tree路由表存储路由
* 使用了新的路由表重写了静态文件系统
