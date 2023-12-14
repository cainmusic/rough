# 路由树设计

一个path的基本结构大概为`/aaa/bbb/ccc`或者`/aaa/bbb/ccc/`

按一般目录的做法，我们大概有：

```
aaa
└── bbb
    └── ccc
```

这样的结构。

不过，gin路由树采用radix tree存储路由，存储结构大致如下：

```
/aaa/bbb/ccc
└── /
```

上层节点代表`/aaa/bbb/ccc`，下层节点`/`联合父节点路径一齐代表`/aaa/bbb/ccc/`。

这种在父节点共用最长相同前缀的结构就是radix tree。

# 通配符

复杂之处在于路由通配符，让我们看三个典型路由：

```
/hello/world
/user/u:id
/static/*file
```

上面三个路由分别是：静态路由、param路由、catchAll路由。

param路由可以获取url中的参数，而catchAll路由则可以匹配后续全部内容。

很明显，param和catchAll是特别的节点，需要从路径中单独提取出来。

比如，我们有`/user/u:id/profile`这样一个url，存储结构应该如下：

```
/user/u
└── :id
    └── /profile
```

# radix tree基本逻辑

radix tree本身的基本逻辑是一个循环或者递归：walk。

这个walk主要使用两个参数：path表示当前路径，n表示当前节点。

n本身有个path，跟当前的path比较，获取最长公共前缀。

## 例1

例如，当前系统里只有一个路由`/hello/world`：

```
/hello/world
```

我们要加入一个新的路由`/hello/china`，我们先查找`/hello/world`和`/hello/china`的最长公共前缀，应为`/hello/`。

我们发现这个部分是`/hello/world`的一部分，于是我们先对其进行分裂，分为`/hello/`和`world`。

```
/hello/
└── world
```

由于我们刚刚利用最长公共前缀对`/hello/world`进行了分裂，明确知道`world`和`china`没有公共前缀。

所以可以直接把`china`加入`/hello/`的子节点：

```
/hello/
├── world
└── china
```

写一个类go的伪代码：

``` go
func walk(n *node, path string) {
	i := longestCommonPrefix(path, n.path)

	if i < len(n.path) {
		// 分裂当前节点
		father, child := split(n, i)

		// 使用path剩下的部分新建一个节点，并用作father的子节点
		newChild := NewChild(path[i:])
		father.addChild(newChild)
	}
}
```

## 例2

继续，如果当前系统里只有一个路由`/hello/world`：

```
/hello/world
```

我们要加入一个新的路由`/hello`，还是先找最长公共前缀，应为`/hello`。

这个部分是`/hello/world`的一部分，我们对其进行分裂：

```
/hello
└── /world
```

这时候需要注意，原本`/hello/world`是有一个处理器的，在gin中，存储在`node.handlers`中。

在此，我们引入一个`[!]`符号表示其节点拥有处理器，并标识到上面的结构中：

```
我们将

/hello/world[!]

分裂为

/hello
└── /world[!]
```

可以看出原本`/hello/world`是一个有处理器的节点，分裂之后，父节点`/hello`不再有处理能力。

之后，我们发现当前的path（`/hello`）在截取最长公共前缀`/hello`之后，不在有内容，所以path到当前节点为止。

形成的结构为：

```
/hello[!]
└── /world[!]
```

至此`/hello`节点获得了处理器。

伪代码：

``` go
func walk(n *node, path string, handlers []handler) {
	i := longestCommonPrefix(path, n.path)

	if i < len(n.path) {
		// 分裂当前节点
		father, child := split(n, i)

		path = path[i:]
		if path == "" {
			father.handlers = handlers
			return
		}

		// 使用path剩下的部分新建一个节点，并用作father的子节点
		newChild := NewChild(path)
		father.addChild(newChild)
	}
}
```

>注：我们同样标识一下【例1】中的最终结构如下：
>```
>/hello/
>├── world[!]
>└── china[!]
>```
>可以看到父节点是没有处理器的，两个字节点分别有一个处理器。

## 例3

继续，如果我们当前只有一个路由`/hello`，此时结构为：

```
/hello[!]
```

我们要加入一个路由`/hello/china`，还是先找最长公共前缀，应为`/hello`。

但此时我们当前节点并不需要进行分裂，而是我们需要直接将path的剩余部分加入当前节点的子节点。

```
/hello[!]
└── /china[!]
```

伪代码：

``` go
func walk(n *node, path string, handlers []handler) {
	i := longestCommonPrefix(path, n.path)

	if i < len(n.path) {
		// 分裂当前节点
		father, child := split(n, i)
		n = father
	}

	path = path[i:]
	if path == "" {
		n.handlers = handlers
		return
	}

	// 使用path剩下的部分新建一个节点，并用作n的子节点
	newChild := NewChild(path)
	n.addChild(newChild)
}
```

## 例4

在【例3】的基础上，即：

```
/hello[!]
└── /china[!]
```

中继续加入路由`/hello/shanghai`，我们从父节点`/hello`开始，寻找最长公共前缀，应为`/hello`。

之后我们获得path的剩余部分`/shanghai`。

一般情况下你可能觉得是不是直接将其插入`/hello`的子节点就可以了？

这样？

```
/hello[!]
├── /china[!]
└── /shanghai[!]
```

错！

很明显，我们前面提到，所有节点都是子节点的前置的最长公共前缀。

`/china`和`/shanghai`显然还有最长公共前缀`/`。

同样需要对`/china`进行分裂，得到：

```
/hello[!]
└── /
    ├── china[!]
    └── shanghai[!]
```

那么，问题来了。

我们要查询当前path跟`/hello`的子节点是否有公共前缀，是不是要对子节点进行遍历，看跟哪个节点有公共前缀？

在回答这个问题前，我们先思考下，最多可以跟几个子节点有公共前缀？

答案是1个。

简单的解释是：如果跟超过一个子节点有公共前缀，那么这几个子节点本身就应该共用这个部分。

一句话解释，还是用前面那句：所有节点都是子节点的前置的最长公共前缀。

举个例子，你要在下面的结构里插入`/ac`：

```
/
├── abc
└── abd
```

你发现`ac`和`abc`以及`abd`都有公共前缀`a`，这说明`abc`和`abd`本身就有前缀需要合并：

```
/
└── ab
    ├── c
    └── d
```

并且这个需要合并的部分是`ab`而不是`a`。

继续加入`ac`得到（暂不考虑子节点顺序问题）：

```
/
└── a
    ├── c
    └── b
        ├── c
        └── d
```

搞清楚上面这个问题之后，我们再来思考一下，很简单，我们直接说结论。

最长公共前缀是多长我们不清楚，但最短长度肯定是1，因为如果是0说明没有公共前缀。

而前面我们知道：最多可以跟1个子节点有公共前缀。

也就是说，当前path，最多与1个子节点，有最少1个长度的公共前缀。

这个结论读起来有点绕，让我们直接看应用。

我们把父节点的所有子节点的路径的首字符拿来，存成一个字符串，比如：

```
/
├── abc
├── def
└── ghi
```

我们拿父节点的所有子节点的首字符组成`adg`，存在父节点上。（gin中存为`node.indices`）

此时，我们插入新的路由`/ask`，我们不再需要遍历`/`的每个字节点。

而是只需要遍历`indices`的每个字符，看是否是`ask`的首字符，即可判断跟哪个字节点有公共前缀。

伪代码：

``` go
path := "ask"
for i, ic := range []byte(n.indices) {
	if ic == path[0] {
		walk(n.children[i], path)
		return
	}
}
// 没找到有公共前缀的子节点，直接成为新的字节点
newChild := NewChild(path)
n.addChild(newChild)
```

当然了，这里有个要求，`indices`的字符顺序和`children`中的对应子节点的顺序要保持一致。

这就是：当前path，最多与1个子节点，有最少1个长度的公共前缀。

暂不考虑indices的构建，本节的伪代码最终为：

``` go
func walk(n *node, path string, handlers []handler) {
	i := longestCommonPrefix(path, n.path)

	if i < len(n.path) {
		// 分裂当前节点
		father, child := split(n, i)
		n = father
	}

	if i < len(path) {
		path = path[i:]
		c := path[0]

		// 这里分两个情况：
		// 如果前面分裂过，可以直接加子节点
		// 如果前面没分裂过，通过indices判断是否有含有公共前缀的子节点
		// 没有就直接加子节点
		// 有就walk下一轮
		// 当然，分裂的情况实际上可以并入第二种情况

		for i, ic := range []byte(n.indices) {
			if c == ic {
				n = n.children[i]
				walk(n, path, handlers)
			}
		}

		newChild := NewChild(path)
		n.addChild(newChild)
	}

	// 对应path截取公共前缀后为空，落在当前节点
	if n.handlers != nil { // handlers不为空，表示重复注册该路由
		panic("already registered")
	}
	n.handlers = handlers
	return
}
```

在本节的结束，思考一个问题：

如果我们按不同的顺序注册`/hello`、`/hello/china`、`/hello/shanghai`三个路由，结果是否一样？

现在我们可以很快的回答，是一样的，一组路由无论按什么顺序注册，结果都应是一样的。

当然了，仍然是不考虑子节点顺序的。

多说一句，也不考虑注册失败的情况，或者说，如果失败，那么以任何顺序注册都会失败，只不过失败时路由树的状态可能不同，但那不重要。

## 小结

通过上面4个例子，我们最终完成了生成radix tree的核心方法walk的大致逻辑。

后面是重头戏，加通配符。

# 在radix tree上加通配符

路由树主体由radix tree构成，但由于路由需要处理param和catchAll两类通配符，需要对radix tree进行改造。

我们之前说radix tree的所有节点都是子节点的前置的最长公共前缀。

加了通配符之后，这个描述应该改为：所有节点都是子节点的前置的最长公共前缀或一个通配符。

那么怎么改造呢，还是让我们看例子。

## 例1

首先注册一个`/user/u:id/profile`的路由。

我们暂时不考虑怎么实现，先畅想一下结果，大概应该是：

```
/user/u
└── :id
    └── /profile
```

一般来说，单一路由应该存为一个节点，但由于其中有通配符，通配符要单独成为节点。

这样做不光是为了存储可视，在应用中对url进行匹配的时候也可以快速找到通配符匹配部分获取数据。

这个操作就是无法找到公共前缀而插入节点的时候的基本操作了。

对于一个path，但凡可以找到匹配节点就使用该分支，匹配不到的情况就用上面的操作插入剩余的path。

## 例2

当前我们已经注册了一个`/user/`的路由，结构如下：

```
/user/[!]
```

之后注册一个`/user/:id/profile`的路由，首先匹配`/user/`，剩余path为`:id/profile`。

之后按【例1】的逻辑走，在`/user/`节点上插入path，即`:id/profile`，得到：

```
/user/[!]
└── :id
    └── /profile[!]
```

而如果我们已经注册了一个`/user`的路由，结构如下：

```
/user[!]
```

之后注册一个`/user/:id/profile`的路由，首先匹配`/user`，剩余path为`/:id/profile`。

之后按【例1】的逻辑走，在`/user`节点上插入path，即`/:id/profile`，得到：

```
/user[!]
└── /
    └── :id
        └── /profile[!]
```

这里分享一个gin的bug，至少目前`v1.9.1`版本还没修复。

``` go
func main() {
	r := gin.Default()

	r.GET("/static/", func(c *gin.Context) { c.String(200, "static") })
	r.GET("/static/*file", func(c *gin.Context) { c.String(200, "static file") })

	r.Run()
}
```

上面的代码会报错：

```
panic: runtime error: index out of range [0] with length 0
```

这个bug是有catchAll通配符的特异性导致的。

param通配符仅匹配`:paramname`部分，获取对应的值。

而catchAll通配符虽然写作`*paramname`，但其构建路由的时候会向前匹配一位`/`。

因为catchAll通配符通常是为了匹配路径而存在的，catchAll通配符在gin中的经典应用就是配置静态文件服务器。

参考gin项目的`routergroup.go`文件：

``` go
func (group *RouterGroup) StaticFS(relativePath string, fs http.FileSystem) IRoutes {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static folder")
	}
	handler := group.createStaticHandler(relativePath, fs)
	urlPattern := path.Join(relativePath, "/*filepath")

	// Register GET and HEAD handlers
	group.GET(urlPattern, handler)
	group.HEAD(urlPattern, handler)
	return group.returnObj()
}
```

一旦你使用`Static`相关函数配置静态文件服务，最后都会调用到上面的方法。

其中用你传入的`relativePath`和`/*filepath`组合为最终的url：`relativePath/*filepath`。

假如你传入的路径是`/static`，则最终url为`/static/*filepath`。

这个路由会匹配所有以`/static/`开头的url，并将后面的所有内容赋值到`filepath`。

没错，是所有，包括后面的`/`，比如`html/group1/page1.html`，也就是说可以通过`filepath`访问到子目录。

但上面描述的内容实际上有一个错误，你以为`filepath`保存的内容是`html/group1/page1.html`。

实际上是`/html/group1/page1.html`。

catchAll通配符会尝试向前多匹配一个`/`，如果你的路由中没有这个`/`，会报错。

这个特性的特异之处导致gin中有一个bug，就是当你已经注册了`/static/`路由之后，再注册`/static/*file`的时候，我们会在`/static/`节点上插入`*filepath`而不是`/*filepath`，这导致在程序判断这是一个catchAll路由后，会去向前匹配一位`/`，这时`i--`后，变成了负数，就导致了`index out of range`的错误。

你可能会觉得报错是对的啊，那么你需要注意分清报错和bug产生的panic。

```
i--
if path[i] != '/' {
	panic("no / before catch-all in path '" + fullPath + "'")
}
```

我说的报错产生在`panic("no / before catch-all in path '" + fullPath + "'")`。

而bug产生的panic产生在`i--`变为负数后查询`if path[i] != '/'`时。

简单处理的话，这里应该对`i`的值进行判断，然后主动panic抛出可以让人领悟的报错。

当然，为了解决这个问题本身，可以将上面的代码修改为：

``` go
func main() {
	r := gin.Default()

	r.GET("/static", func(c *gin.Context) { c.String(200, "static") })
	r.GET("/static/*file", func(c *gin.Context) { c.String(200, "static file") })

	r.Run()
}
```

第一个路由不要加末尾的`/`，就可以规避这个bug。

## 小结

通配符的加入就是把通配符从url中找出来，并形成一个单一节点再想办法加入到radix tree中。

然后同时再处理掉通配符的一些特异性问题，比如：

* 一个父节点只能有一个通配符子节点
* catchAll通配符前面必须有`/`
* catchAll通配符不能有兄弟节点也不能有子节点
* param通配符可以有兄弟节点，但必须位于兄弟节点切片的最末端，避免匹配时遮盖前面的节点
* url中的一个segment中（指两个`/`之间）只能有一个param通配符

等等。（如果后续有想到新的会在这里补充）

# 实现

最后让我们来实现一个路由树。

参考`./tree.go`



注意：

通配符虽然是`:`和`*`，但通配符存储节点的path的开头分别是`:`和`/*`。

另外`*`通配符必然有一个path为空的父节点，可以理解为`*`通配符开头的`/`就是从这个父节点抢来的。

节点的indices中必然没有`*`而是`/`，但这里有个问题，indices中的`/`不代表对应的是一个通配符子节点，也可能只是个static节点以`/`开头。

并且没法在当前节点通过wildChild字段确认该节点是否为通配符节点，因为通配符节点会有一个空父节点，虽然该节点类型为catchAll，但其不会将其父节点的wildChild置为true。

catchAll通配符只能在setPath（setChild）中处理，其他方法不可以创建catchAll通配符。

新增两点：

1. 由于catchAll通配符向前匹配一位`/`，可能导致`/*filepath`这样的通配符需要直接落在根节点上，但这样会改变根节点的类型，这是不可以的，所以catchAll通配符确实需要引入一个空父节点

2. gin的tree.go中调用insertChild的地方基本都新建了一个节点作为当前节点，过去我的认知是，这会导致如果insertChild的path是以`/*`开头的时候，无法追溯父节点设置wildChild字段，但如果考虑到前面第1点，实际上这个问题被无形中解决了

后续，还是把setPath改回setChild，并在进入方法前在需要创建新节点的时候提前创建。

我的setPath修改实际上引入了一个新的问题，我在setPath中再创建新的节点，但实际上第一个setPath是由root节点调用的，创建新节点会导致root节点变空，需要进行对root节点进行特殊的处理，所以这样确实不可以






区别：

1. 将addRoute拆成了addRoute和walk，用walk递归替代原本的walk循环
2. 进入walk第一步判断当前节点是否catchAll，catchAll节点不能做任何操作，直接panic
3. 寻找了最大公共前缀之后第一步就是判断这个前缀是否为空，为空则直接panic