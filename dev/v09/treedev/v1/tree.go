package tree

import (
	"fmt"
)

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// 最长公共前缀
func longestCommonPrefix(a, b string) int {
	i := 0
	max := min(len(a), len(b))
	for i < max && a[i] == b[i] {
		i++
	}
	return i
}

type nodeType uint8

const (
	static nodeType = iota
	root
	//param
	//catchAll
)

type node struct {
	path string

	nType nodeType

	// 存储children的path首字母，可以快速判断和新增节点是否有公共前缀
	// 压缩树的原理是通过最大公共前缀来对树进行规划修改的
	// indices存储children的path首字母，这样判断path的当前byte在indices中是否存在，就可以快速判断是否有公共前缀
	indices string

	// 节点优先级，记录自己和子节点被注册的次数，并根据优先级进行排序，确保多次被注册的节点考前
	// 这样可以保证在查询的时候路由尽量更快被找到
	priority uint32

	// 子节点
	children []*node

	fullPath string

	// 占位字段，在gin和rough中可以改为handler函数的切片
	tokens []int
}

type nodeRes struct {
	fullPath string
	tokens   []int
}

func (n *node) addRoute(path string, tokens []int) {
	fullPath := path
	n.priority++

	// 空树的根节点，通常由new(node)或&node{}创建
	if len(n.path) == 0 && len(n.children) == 0 {
		n.insertChild(path, fullPath, tokens)
		n.nType = root
		return
	}

	// 父节点全路径在当前全路径中的index
	parentFullPathIndex := 0

	/*
	   主循环的逻辑就是不断比较当前节点和插入节点是否有公共前缀
	   这里的算法主要是通过剪切公共前缀不断缩短插入节点，最终将该节点的所有父节点连接到它，就是完整path

	   先求公共前缀，有下列情况：
	   1. 当前节点与插入节点无公共前缀，那么分两个情况
	    1. 当前节点为空，直接成为当前节点的子节点
	    2. 当前节点不为空，则分裂当前节点为一个空的父节点，并把当前节点插作子节点，新插入节点为空节点的子节点

	   2. 当前节点与插入节点有公共前缀，那么分四个情况
	    1. 当前节点和插入节点相同，重复注册，报错
	    2. 当前节点是插入节点的子前缀，将插入节点剪切当前节点的部分，形成新节点，结合当前节点的children判断是进入下一循环还是插入当前children
	    3. 插入节点是当前节点的子前缀，将当前节点多出来的部分继承当前节点的信息为子节点，将插入节点信息写入当前节点为父节点
	    4. 当前节点和插入节点不互相包含，但有公共子前缀，将当前节点分裂，多出来的部分继承当前节点信息为子节点，父节点为空，将插入节点也作为空父节点的children

	   优化：
	   前面的操作可以抽象为两个基本操作：
	   1. 是否分裂当前节点，如何处理当前节点的信息
	   2. 剪切插入节点的公共前缀部分，看是否有盈余
	    1. 无则在当下进行覆盖
	    2. 有则在子节点中发现新的“当前节点”，找到了就进入下一循环，没找到就插入当前节点的children

	   额外加一个路由重复，对应前面2.1
	*/
walk:
	for {
		i := longestCommonPrefix(path, n.path)

		// 如果i小于当前节点的长度，分裂当前节点
		flagSplit := false
		if i < len(n.path) {
			flagSplit = true

			child := node{
				// 子节点由i位置起，当前节点则收缩为path[:i]或n.path[:i]
				path:  n.path[i:],
				nType: static,
				// 子节点继承当前节点信息
				indices:  n.indices,
				priority: n.priority - 1, // 函数入口进行了++，此处因为分裂，将优先级还回去
				children: n.children,
				fullPath: n.fullPath,
				tokens:   n.tokens,
			}

			n.children = []*node{&child}
			n.indices = string([]byte{n.path[i]})
			n.path = path[:i] // 理论上也可以用n.path[:i]，TODO：测试验证
			n.fullPath = fullPath[:parentFullPathIndex+i]
			n.tokens = nil
		}

		// 插入节点处理
		if i < len(path) {
			path = path[i:]
			c := path[0]

			// 如果未发生分裂，快速判断子节点是否有公共前缀，有则进入下一循环
			if !flagSplit {
				for i, max := 0, len(n.indices); i < max; i++ {
					if c == n.indices[i] {
						parentFullPathIndex += len(n.path)
						i = n.incrementChildPrio(i)
						n = n.children[i]
						continue walk
					}
				}
			}

			// 无法进入下一循环则在当前节点插入
			n.indices += string([]byte{c})
			child := &node{
				fullPath: fullPath,
			}
			n.addChild(child)
			n.incrementChildPrio(len(n.indices) - 1)
			n = child

			n.insertChild(path, fullPath, tokens)
			return
		}

		if n.tokens != nil {
			panic("tokens are already registered for path '" + fullPath + "'")
		}

		n.fullPath = fullPath
		n.tokens = tokens
		return
	}
}

// 一般radix tree如下处理就可以了，但若引入通配符则需要处理children的排序，所以将addChild方法独立出来
func (n *node) addChild(child *node) {
	n.children = append(n.children, child)
}

// 根据priority重建排序，复制gin/tree.go对应方法
func (n *node) incrementChildPrio(pos int) int {
	cs := n.children
	cs[pos].priority++
	prio := cs[pos].priority

	// Adjust position (move to front)
	newPos := pos
	for ; newPos > 0 && cs[newPos-1].priority < prio; newPos-- {
		// Swap node positions
		cs[newPos-1], cs[newPos] = cs[newPos], cs[newPos-1]
	}

	// Build new index char string
	if newPos != pos {
		n.indices = n.indices[:newPos] + // Unchanged prefix, might be empty
			n.indices[pos:pos+1] + // The index char we move
			n.indices[newPos:pos] + n.indices[pos+1:] // Rest without char at 'pos'
	}

	return newPos
}

// 一般radix tree如下处理就可以了，但若引入通配符则需要继续对path进行分割，所以将insertChild方法独立出来
func (n *node) insertChild(path string, fullPath string, tokens []int) {
	// 不含通配符的情况
	n.path = path
	n.fullPath = fullPath
	n.tokens = tokens
}

func (n *node) getRoute(path string) nodeRes {
	var res nodeRes
walk:
	for {
		prefix := n.path
		if len(path) > len(prefix) {
			if path[:len(prefix)] == prefix {
				path = path[len(prefix):]
				idxc := path[0]
				for i, c := range []byte(n.indices) {
					if c == idxc {
						n = n.children[i]
						continue walk
					}
				}
			}
		}

		if path == prefix {
			if res.tokens = n.tokens; res.tokens != nil {
				res.fullPath = n.fullPath
				return res
			}
		}

		return res
	}
}

type debugNode struct {
	floor int
	node  *node
}

func (n *node) debugPrint() {
	l := &[]debugNode{}
	f := 0
	n.debugRead(l, f)
	debugTreePrint(l)
}

func (n *node) debugRead(l *[]debugNode, f int) {
	*l = append(*l, debugNode{
		floor: f,
		node:  n,
	})
	f++
	for _, c := range n.children {
		c.debugRead(l, f)
	}
	f--
}

func debugTreePrint(l *[]debugNode) {
	fmt.Println("prio,floor,path  ,full           ,tokens")
	for _, d := range *l {
		fmt.Printf(
			"%d   ,%d    ,%s,%s,%v\n",
			d.node.priority,
			d.floor,
			endSpaces(d.node.path, 6),
			endSpaces(d.node.fullPath, 15),
			d.node.tokens)
	}
}

func endSpaces(s string, n int) string {
	for len(s) < n {
		s = s + " "
	}
	return s
}
