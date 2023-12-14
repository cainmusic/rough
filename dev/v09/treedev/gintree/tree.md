# gin tree

在学习gin tree的过程中，逐渐掌握了很多文档里没有的细节，并形成了当前目录的上一级目录中v1和v2的代码。

之后我把gin tree的测试迁移过来配合我自己的测试功能，完成了对v2的测试，并修改了一些bug。

# v2

v2和gin tree的实现基本相同了。

但因为判断逻辑不同，代码优化不同，所以部分测试没有完全按gin tree的测试用例实现。

但该panic的都panic了，只是报错信息的详尽程度不同。

# gin tree的修改

同时，gin tree也不是完美的。

我对当前目录下的gin tree进行了一些修改，下面将列出：

## 改动

在当前目录，通过`diff tree.go gin/tree.go`命令可以查看代码改动。

## bug: index out of range

gin tree有一个bug代码如下：

```
package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/static/", func(c *gin.Context) { c.String(200, "static") })
	r.GET("/static/*file", func(c *gin.Context) { c.String(200, "static file") })

	r.Run()
}
```

运行会报：

```
panic: runtime error: index out of range [0] with length 0
```

是因为下面这段源码有问题：

```
		if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
			pathSeg := strings.SplitN(n.children[0].path, "/", 2)[0]
			panic("catch-all wildcard '" + path +
				"' in new path '" + fullPath +
				"' conflicts with existing path segment '" + pathSeg +
				"' in existing prefix '" + n.path + pathSeg +
				"'")
		}
```

其中第二行`(n.children[0].path`应该改为`n.path`。

这个bug我提交issue了：[链接](https://github.com/gin-gonic/gin/issues/3796)，但目前没有反馈。

## 构建结果不同

（使用`github.com/cainmusic/gtable`美化路由输出。）

一般来说改变两个路由的注册顺序，其生成的路由树结构应该保持一致。

但gin tree在遇到`catch-all`通配符路由时会遇到下列问题：

先后注册`/static`和`/static/*filepath`（此处handlers代表注册顺序）：

```
+----+-----+----------+-----------------+--------+-----+---------+-------+
|prio|floor|path      |fullPath         |handlers|nType|wildChild|indices|
+----+-----+----------+-----------------+--------+-----+---------+-------+
|2   |0    |/static   |/static          |1       |1    |false    |/      |
|1   |1    |          |/static/*filepath|<nil>   |0    |false    |/      |
|1   |2    |          |/static/*filepath|<nil>   |3    |true     |       |
|1   |3    |/*filepath|/static/*filepath|2       |3    |false    |       |
+----+-----+----------+-----------------+--------+-----+---------+-------+
```

先后注册`/static/*filepath`和`/static`（此处handlers代表注册顺序）：

```
+----+-----+----------+-----------------+--------+-----+---------+-------+
|prio|floor|path      |fullPath         |handlers|nType|wildChild|indices|
+----+-----+----------+-----------------+--------+-----+---------+-------+
|2   |0    |/static   |/static          |2       |1    |false    |/      |
|1   |1    |          |/static/*filepath|<nil>   |3    |true     |       |
|1   |2    |/*filepath|/static/*filepath|1       |3    |false    |       |
+----+-----+----------+-----------------+--------+-----+---------+-------+
```

不同的注册顺序产生了不同的路由树。

参考`./gin/tree_my_test.go`中的`pathCase0301`和`pathCase0302`。

之后再测试可得：

先后注册`/static`和`/static/*filepath`（此处handlers代表注册顺序）：

```
+----+-----+----------+-----------------+--------+-----+---------+-------+
|prio|floor|path      |fullPath         |handlers|nType|wildChild|indices|
+----+-----+----------+-----------------+--------+-----+---------+-------+
|2   |0    |/static   |/static          |1       |1    |false    |/      |
|1   |1    |          |/static/*filepath|<nil>   |3    |true     |       |
|1   |2    |/*filepath|/static/*filepath|2       |3    |false    |       |
+----+-----+----------+-----------------+--------+-----+---------+-------+
```

先后注册`/static/*filepath`和`/static`（此处handlers代表注册顺序）：

```
+----+-----+----------+-----------------+--------+-----+---------+-------+
|prio|floor|path      |fullPath         |handlers|nType|wildChild|indices|
+----+-----+----------+-----------------+--------+-----+---------+-------+
|2   |0    |/static   |/static          |2       |1    |false    |/      |
|1   |1    |          |/static/*filepath|<nil>   |3    |true     |       |
|1   |2    |/*filepath|/static/*filepath|1       |3    |false    |       |
+----+-----+----------+-----------------+--------+-----+---------+-------+
```

参考`./tree_my_test.go`中的`pathCase0301`和`pathCase0302`。

如此，保证了不同构建顺序产生的结果相同。

另外，由于本次改动导致了indices字段的初始化改动，间接影响了getValue时nodeValue的tsr，也进行了修正。

# 其他细节

gin tree有很多实现中才能逐步发现的细节：

1. catch-all通配符会向前掠夺一位"/"，并会导致一个空节点的产生，这个空节点和catch-all节点本身构成了一个catch-all节点组合
    * 一般根节点的path至少会包含一个"/"，但如果在根节点上注册一个catch-all通配符，则会产生两个空节点（包括根节点）
2. param通配符的首字符":"是不写入父节点的indices的，所以不能通过查询indices的方法匹配子节点
3. 但是catch-all掠夺来的"/"是会进入父节点的indices的，不过这不影响indices的排序
    * indices中的"/"有两种含义
    * 如果只是一个static节点，则可以参与排序
    * 如果代表了一个catch-all节点，由于catch-all节点不能有兄弟节点，所以indices中的"/"是唯一的，不影响排序
4. 待补充

# 部署

进行必要修改，并将当前目录的tree.go（修改后的gin tree）部署到rough中。
