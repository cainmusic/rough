package tree

import (
	"bytes"
	"net/url"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

func StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func BytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

var (
	strColon = []byte(":")
	strStar  = []byte("*")
	strSlash = []byte("/")
)

type Param struct {
	Key   string
	Value string
}

type Params []Param

// utils

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func longestCommonPrefix(a, b string) int {
	i := 0
	max := min(len(a), len(b))
	for i < max && a[i] == b[i] {
		i++
	}
	return i
}

// 从gin项目tree.go文件复制而来
func (n *node) addChild(child *node) {
	if n.wildChild && len(n.children) > 0 {
		wildcardChild := n.children[len(n.children)-1]
		n.children = append(n.children[:len(n.children)-1], child, wildcardChild)
	} else {
		n.children = append(n.children, child)
	}
}

func countParams(path string) uint16 {
	var n uint16
	s := StringToBytes(path)
	n += uint16(bytes.Count(s, strColon))
	n += uint16(bytes.Count(s, strStar))
	return n
}

// node

type nodeType uint8

const (
	static nodeType = iota
	root
	param
	catchAll
)

type node struct {
	path      string
	indices   string
	wildChild bool // 快速了解children中有没有wildcard
	nType     nodeType
	priority  uint32
	children  []*node
	handlers  any
	fullPath  string
}

// 从gin项目tree.go文件复制而来
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

func (n *node) addRoute(path string, handlers any) {
	fullPath := path
	n.priority++

	// 空树
	if len(n.path) == 0 && len(n.children) == 0 {
		n.nType = root
		n.setChild(path, handlers, fullPath, 0)
		return
	}

	parentFullPathIndex := 0

	n.walk(path, handlers, fullPath, parentFullPathIndex)
}

func (n *node) walk(path string, handlers any, fullPath string, parentFullPathIndex int) {
	// 首先判断catchAll和param节点的情况

	// catchAll节点不能做任何操作，不能分裂，不能加子节点，也不能重新注册
	if n.nType == catchAll {
		panic("catch-all node cannot do anything more, no split, no children, no siblings, no re-register")
	}

	i := longestCommonPrefix(path, n.path)
	//fmt.Println(i, path)

	// 进入walk方法一般要确保path与n.path有公共前缀，否则应该加子节点
	// 还有一种特别的情况就是当前节点为空节点，如果是空节点，则接下来必然匹配到catchAll节点
	// 所以只要i为0就应该panic
	if i == 0 {
		panic("path and n.path should have same prefix, or current node is father for catch-all")
	}

	// 公共前缀小于当前节点的path，分裂当前节点
	if i < len(n.path) {
		child := &node{
			// path分裂
			path: n.path[i:],
			// 可分裂的必不是通配符，新建子节点必不是root
			nType: static,
			// 子节点继承分裂前节点的其他信息
			wildChild: n.wildChild,
			indices:   n.indices,
			children:  n.children,
			handlers:  n.handlers,
			priority:  n.priority - 1, // 分裂后priority的增加应该在分裂得到的父节点上，归还1个priority
			fullPath:  n.fullPath,
		}

		// 分裂出来的新的父节点只有分裂出来的一个新的子节点
		n.children = []*node{child}
		// 原代码涉及到特殊的unicode适应，这里暂不考虑
		n.indices = string([]byte{n.path[i]})
		n.path = path[:i]
		n.handlers = nil
		n.wildChild = false
		n.fullPath = fullPath[:parentFullPathIndex+i]
	}

	/*
		此时path有若干种可能：
		1. 以param通配符为起点
		:param/xxx
		2. 以一般字符串为起点，后接param通配符
		xxx:param/xxx
		3. 以catchAll通配符为起点
		*file
		4. 以/为起点，后接catchAll通配符
		/*file
		5. 以一般字符串为起点，后接/，再接catchAll通配符
		xxx/*file
		6. 一般字符串
		xxx
		7. 以/开头的一般字符串
		/xxx
	*/

	// 公共前缀小于path，需要使path的后段（公共前缀后面的部分）成为当前节点的子节点
	// 这个子节点可能是新的，也可能与已存在的子节点结合
	if i < len(path) {
		// path抛弃公共部分
		// 注意，此处path后段不可能为空
		path = path[i:]
		c := path[0]

		// 前面已经判断n.nType为catchAll或param的情况，接下来：
		// 1. 判断path以catchAll和param为起点的情况
		// 2. 查找可以匹配的子节点
		// 3. 没找到匹配的子节点，加子节点（不再在walk中调用addChild，仅调用setChild，在setChild中调用addChild，进行解耦）
		// 4. 落当前节点

		// 1

		// path以*开头，无法向前回溯一位"/"，直接panic
		if c == '*' {
			panic("catch-all node must have free '/' before '*'")
		}

		parentFullPathIndex += len(n.path)

		// path以/*开头，当前无法判断，查看有无可以walk的子节点，没有就交给setChild处理
		// path以xxx/*file开头，当前无法判断，查看有无可以walk的子节点，没有就交给setChild处理

		// path以:开头
		if c == ':' {
			if n.wildChild {
				// 如果n.wildChild为true，需要判断和通配符子节点是否相同
				// 相同则walk，不相同则panic
				n = n.children[len(n.children)-1]
				n.priority++
				// path长度小于n.path，param不一致
				if len(n.path) > len(path) ||
					// path前段与n.path不一致，param不一致
					n.path != path[:len(n.path)] ||
					// path前段与n.path一致，但后续不为分隔符"/"，param不一致
					// 例：":abc"与":abcd"冲突，":abc"与":abc/"不冲突
					(len(path) > len(n.path) && n.path == path[:len(n.path)] && path[len(n.path)] != '/') {
					panic("param node cannot have wildcard siblings")
				}
				// 判断path和n.path必有相同param通配符，确保进入walk后最大公共前缀长度不为0
				n.walk(path, handlers, fullPath, parentFullPathIndex)
				return
			} else {
				// 如果n.wildChild为false，加子节点
				n.setChild(path, handlers, fullPath, parentFullPathIndex)
				return
			}
		}

		// path以xxx:param开头，查看有无可以walk的子节点，没有就交给setChild处理

		// 2

		// 这是一个快捷操作，如果不用也不会出错，而是会进入后面的循环
		// 此处的快捷操作在于，当nType为param时，如果有合法的后段path，必定是以"/"开头
		// 同时如果param有子节点，则必定只有一个子节点且以"/"开头
		// 注意这里的三个判断缺一不可
		// 首先，必须是param通配符
		// 其次，如果c不为"/"，后续会报通配符冲突，并且c也必须为"/"才能合法走下一步
		// 最后，len(n.children)必须为1才可以快捷操作，否则可能为0，没有子节点则无法操作
		if n.nType == param && c == '/' && len(n.children) == 1 {
			n = n.children[0]
			n.priority++
			// 前面已经说明，至少有相同字符"/"确保进入walk后最大公共前缀长度不为0
			n.walk(path, handlers, fullPath, parentFullPathIndex)
			return
		}

		// 利用indices快速判断path可以与哪个子节点结合，进入下一个walk
		for i := 0; i < len(n.indices); i++ {
			if c == n.indices[i] {
				i = n.incrementChildPrio(i)
				n = n.children[i]
				// 已通过indices确保进入walk后最大公共前缀长度不为0
				n.walk(path, handlers, fullPath, parentFullPathIndex)
				return
			}
		}

		// 前面已经判断path是否可以和某个子节点进行结合
		// 后面将进行插入新子节点操作
		n.indices += string([]byte{c})
		child := &node{
			fullPath: fullPath, // TODO，此处是否使用fullPath存疑，后续看是否改为fullPath[:parentFullPathIndex]
		}
		n.addChild(child)
		n.incrementChildPrio(len(n.indices) - 1)
		n = child

		n.setChild(path, handlers, fullPath, parentFullPathIndex)
		return
	}

	// 对应path截取公共前缀后为空，落在当前节点
	if n.handlers != nil { // handlers不为空，表示重复注册该路由
		panic("already registered")
	}
	n.handlers = handlers
	n.fullPath = fullPath
	return
}

// 从gin项目tree.go文件复制而来
func findWildcard(path string) (wildcard string, i int, valid bool) {
	// Find start
	for start, c := range []byte(path) {
		// A wildcard starts with ':' (param) or '*' (catch-all)
		if c != ':' && c != '*' {
			continue
		}

		// Find end and check for invalid characters
		valid = true
		for end, c := range []byte(path[start+1:]) {
			switch c {
			case '/':
				return path[start : start+1+end], start, valid
			case ':', '*':
				valid = false
			}
		}
		return path[start:], start, valid
	}
	return "", -1, false
}

// setChild方法，进入之前一定要确认当前节点n是可以写入的节点
// 比如还未初始化的root节点，或者path以非通配符开始
func (n *node) setChild(path string, handlers any, fullPath string, parentFullPathIndex int) {
	for {
		// 每次寻找一个通配符
		wildcard, i, valid := findWildcard(path)
		// 没有通配符，跳出
		if i < 0 {
			break
		}

		if !valid {
			panic("only one wildcard per path segment is allowed")
		}

		if len(wildcard) < 2 {
			panic("wildcard must have a name")
		}

		// param通配符处理
		if wildcard[0] == ':' {
			// 前面walk方法中
			// 如果以非:开头，则创建了新节点，此处i必大于0，将param通配符前面的内容赋值到当前节点
			// 如果path以:开头，则没有创建新节点直接调用了setChild，此处i等于0，就不会影响当前节点的值
			if i > 0 {
				n.path = path[:i]
				path = path[i:]
				//n.indices = ":"
			}

			parentFullPathIndex += i // TODO，加了但暂时没用

			child := &node{
				nType:    param,
				path:     wildcard,
				fullPath: fullPath,
			}
			n.addChild(child)
			n.wildChild = true
			n = child
			n.priority++

			// 如果当前的param通配符后面还有内容，则回到循环
			// 这里直接加子节点也可以，因为param后续必定是"/"
			if len(wildcard) < len(path) {
				path = path[len(wildcard):]

				child := &node{
					priority: 1,
					fullPath: fullPath,
				}
				n.indices = "/"
				n.addChild(child)
				n = child
				continue
			}

			// param之后无内容，路由到此为止
			n.handlers = handlers
			return
		}

		// catchAll通配符处理
		if i+len(wildcard) != len(path) {
			panic("catch-all routes are only allowed at the end of the path in path '" + fullPath + "'")
		}

		// 回溯一位，此处无需判断i--后为负，因为这种情况在walk中已经被过滤
		i--
		if path[i] != '/' {
			panic("no '/' before catch-all")
		}

		wildcard = "/" + wildcard

		// catchAll通配符标配一个空节点加一个通配符节点

		// 空节点处理
		// 这里做了一个额外处理来确保不会出现多的空节点
		if i > 0 || n.nType == root {
			n.path = path[:i]

			child := &node{
				wildChild: true,
				nType:     catchAll,
				fullPath:  fullPath,
				priority:  1,
			}
			n.addChild(child)
			n.indices = string('/')
			n = child
		} else {
			n.wildChild = true
			n.nType = catchAll
			n.fullPath = fullPath
		}

		// catchAll节点
		child := &node{
			path:     path[i:],
			nType:    catchAll,
			handlers: handlers,
			priority: 1,
			fullPath: fullPath,
		}
		n.children = []*node{child}

		return
	}

	// 没找到通配符，直接加子节点
	n.path = path
	n.handlers = handlers
	n.fullPath = fullPath
}

type nodeValue struct {
	handlers any
	params   *Params
	tsr      bool
	fullPath string
}

type skippedNode struct {
	path        string
	node        *node
	paramsCount int16
}

func (n *node) getValue(path string, params *Params, skippedNodes *[]skippedNode, unescape bool) (value nodeValue) {
	var globalParamsCount int16

walk: // Outer loop for walking the tree
	for {
		prefix := n.path
		if len(path) > len(prefix) {
			if path[:len(prefix)] == prefix {
				path = path[len(prefix):]

				// Try all the non-wildcard children first by matching the indices
				idxc := path[0]
				for i, c := range []byte(n.indices) {
					if c == idxc {
						//  strings.HasPrefix(n.children[len(n.children)-1].path, ":") == n.wildChild
						if n.wildChild {
							index := len(*skippedNodes)
							*skippedNodes = (*skippedNodes)[:index+1]
							(*skippedNodes)[index] = skippedNode{
								path: prefix + path,
								node: &node{
									path:      n.path,
									wildChild: n.wildChild,
									nType:     n.nType,
									priority:  n.priority,
									children:  n.children,
									handlers:  n.handlers,
									fullPath:  n.fullPath,
								},
								paramsCount: globalParamsCount,
							}
						}

						n = n.children[i]
						continue walk
					}
				}

				if !n.wildChild {
					// If the path at the end of the loop is not equal to '/' and the current node has no child nodes
					// the current node needs to roll back to last valid skippedNode
					if path != "/" {
						for length := len(*skippedNodes); length > 0; length-- {
							skippedNode := (*skippedNodes)[length-1]
							*skippedNodes = (*skippedNodes)[:length-1]
							if strings.HasSuffix(skippedNode.path, path) {
								path = skippedNode.path
								n = skippedNode.node
								if value.params != nil {
									*value.params = (*value.params)[:skippedNode.paramsCount]
								}
								globalParamsCount = skippedNode.paramsCount
								continue walk
							}
						}
					}

					// Nothing found.
					// We can recommend to redirect to the same URL without a
					// trailing slash if a leaf exists for that path.
					value.tsr = path == "/" && n.handlers != nil
					return
				}

				// Handle wildcard child, which is always at the end of the array
				n = n.children[len(n.children)-1]
				globalParamsCount++

				switch n.nType {
				case param:
					// fix truncate the parameter
					// tree_test.go  line: 204

					// Find param end (either '/' or path end)
					end := 0
					for end < len(path) && path[end] != '/' {
						end++
					}

					// Save param value
					if params != nil && cap(*params) > 0 {
						if value.params == nil {
							value.params = params
						}
						// Expand slice within preallocated capacity
						i := len(*value.params)
						*value.params = (*value.params)[:i+1]
						val := path[:end]
						if unescape {
							if v, err := url.QueryUnescape(val); err == nil {
								val = v
							}
						}
						(*value.params)[i] = Param{
							Key:   n.path[1:],
							Value: val,
						}
					}

					// we need to go deeper!
					if end < len(path) {
						if len(n.children) > 0 {
							path = path[end:]
							n = n.children[0]
							continue walk
						}

						// ... but we can't
						value.tsr = len(path) == end+1
						return
					}

					if value.handlers = n.handlers; value.handlers != nil {
						value.fullPath = n.fullPath
						return
					}
					if len(n.children) == 1 {
						// No handle found. Check if a handle for this path + a
						// trailing slash exists for TSR recommendation
						f := n
						n = n.children[0]
						value.tsr = (n.path == "/" && n.handlers != nil) || (n.path == "" && f.indices == "/")
					}
					return

				case catchAll:
					// Save param value
					if params != nil {
						if value.params == nil {
							value.params = params
						}
						// Expand slice within preallocated capacity
						i := len(*value.params)
						*value.params = (*value.params)[:i+1]
						val := path
						if unescape {
							if v, err := url.QueryUnescape(path); err == nil {
								val = v
							}
						}
						(*value.params)[i] = Param{
							Key:   n.path[2:],
							Value: val,
						}
					}

					value.handlers = n.handlers
					value.fullPath = n.fullPath
					return

				default:
					panic("invalid node type")
				}
			}
		}

		if path == prefix {
			// If the current path does not equal '/' and the node does not have a registered handle and the most recently matched node has a child node
			// the current node needs to roll back to last valid skippedNode
			if n.handlers == nil && path != "/" {
				for length := len(*skippedNodes); length > 0; length-- {
					skippedNode := (*skippedNodes)[length-1]
					*skippedNodes = (*skippedNodes)[:length-1]
					if strings.HasSuffix(skippedNode.path, path) {
						path = skippedNode.path
						n = skippedNode.node
						if value.params != nil {
							*value.params = (*value.params)[:skippedNode.paramsCount]
						}
						globalParamsCount = skippedNode.paramsCount
						continue walk
					}
				}
				//	n = latestNode.children[len(latestNode.children)-1]
			}
			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.
			if value.handlers = n.handlers; value.handlers != nil {
				value.fullPath = n.fullPath
				return
			}

			// If there is no handle for this route, but this route has a
			// wildcard child, there must be a handle for this path with an
			// additional trailing slash
			if path == "/" && n.wildChild && n.nType != root {
				value.tsr = true
				return
			}

			if path == "/" && n.nType == static {
				value.tsr = true
				return
			}

			// No handle found. Check if a handle for this path + a
			// trailing slash exists for trailing slash recommendation
			for i, c := range []byte(n.indices) {
				if c == '/' {
					n = n.children[i]
					value.tsr = (len(n.path) == 1 && n.handlers != nil) ||
						(n.nType == catchAll && n.children[0].handlers != nil)
					return
				}
			}

			return
		}

		// Nothing found. We can recommend to redirect to the same URL with an
		// extra trailing slash if a leaf exists for that path
		value.tsr = path == "/" ||
			(len(prefix) == len(path)+1 && prefix[len(path)] == '/' &&
				path == prefix[:len(prefix)-1] && n.handlers != nil)

		// roll back to last valid skippedNode
		if !value.tsr && path != "/" {
			for length := len(*skippedNodes); length > 0; length-- {
				skippedNode := (*skippedNodes)[length-1]
				*skippedNodes = (*skippedNodes)[:length-1]
				if strings.HasSuffix(skippedNode.path, path) {
					path = skippedNode.path
					n = skippedNode.node
					if value.params != nil {
						*value.params = (*value.params)[:skippedNode.paramsCount]
					}
					globalParamsCount = skippedNode.paramsCount
					continue walk
				}
			}
		}

		return
	}
}

// Makes a case-insensitive lookup of the given path and tries to find a handler.
// It can optionally also fix trailing slashes.
// It returns the case-corrected path and a bool indicating whether the lookup
// was successful.
func (n *node) findCaseInsensitivePath(path string, fixTrailingSlash bool) ([]byte, bool) {
	const stackBufSize = 128

	// Use a static sized buffer on the stack in the common case.
	// If the path is too long, allocate a buffer on the heap instead.
	buf := make([]byte, 0, stackBufSize)
	if length := len(path) + 1; length > stackBufSize {
		buf = make([]byte, 0, length)
	}

	ciPath := n.findCaseInsensitivePathRec(
		path,
		buf,       // Preallocate enough memory for new path
		[4]byte{}, // Empty rune buffer
		fixTrailingSlash,
	)

	return ciPath, ciPath != nil
}

// Shift bytes in array by n bytes left
func shiftNRuneBytes(rb [4]byte, n int) [4]byte {
	switch n {
	case 0:
		return rb
	case 1:
		return [4]byte{rb[1], rb[2], rb[3], 0}
	case 2:
		return [4]byte{rb[2], rb[3]}
	case 3:
		return [4]byte{rb[3]}
	default:
		return [4]byte{}
	}
}

// Recursive case-insensitive lookup function used by n.findCaseInsensitivePath
func (n *node) findCaseInsensitivePathRec(path string, ciPath []byte, rb [4]byte, fixTrailingSlash bool) []byte {
	npLen := len(n.path)

walk: // Outer loop for walking the tree
	for len(path) >= npLen && (npLen == 0 || strings.EqualFold(path[1:npLen], n.path[1:])) {
		// Add common prefix to result
		oldPath := path
		path = path[npLen:]
		ciPath = append(ciPath, n.path...)

		if len(path) == 0 {
			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.
			if n.handlers != nil {
				return ciPath
			}

			// No handle found.
			// Try to fix the path by adding a trailing slash
			if fixTrailingSlash {
				for i, c := range []byte(n.indices) {
					if c == '/' {
						n = n.children[i]
						if (len(n.path) == 1 && n.handlers != nil) ||
							(n.nType == catchAll && n.children[0].handlers != nil) {
							return append(ciPath, '/')
						}
						return nil
					}
				}
			}
			return nil
		}

		// If this node does not have a wildcard (param or catchAll) child,
		// we can just look up the next child node and continue to walk down
		// the tree
		if !n.wildChild {
			// Skip rune bytes already processed
			rb = shiftNRuneBytes(rb, npLen)

			if rb[0] != 0 {
				// Old rune not finished
				idxc := rb[0]
				for i, c := range []byte(n.indices) {
					if c == idxc {
						// continue with child node
						n = n.children[i]
						npLen = len(n.path)
						continue walk
					}
				}
			} else {
				// Process a new rune
				var rv rune

				// Find rune start.
				// Runes are up to 4 byte long,
				// -4 would definitely be another rune.
				var off int
				for max := min(npLen, 3); off < max; off++ {
					if i := npLen - off; utf8.RuneStart(oldPath[i]) {
						// read rune from cached path
						rv, _ = utf8.DecodeRuneInString(oldPath[i:])
						break
					}
				}

				// Calculate lowercase bytes of current rune
				lo := unicode.ToLower(rv)
				utf8.EncodeRune(rb[:], lo)

				// Skip already processed bytes
				rb = shiftNRuneBytes(rb, off)

				idxc := rb[0]
				for i, c := range []byte(n.indices) {
					// Lowercase matches
					if c == idxc {
						// must use a recursive approach since both the
						// uppercase byte and the lowercase byte might exist
						// as an index
						if out := n.children[i].findCaseInsensitivePathRec(
							path, ciPath, rb, fixTrailingSlash,
						); out != nil {
							return out
						}
						break
					}
				}

				// If we found no match, the same for the uppercase rune,
				// if it differs
				if up := unicode.ToUpper(rv); up != lo {
					utf8.EncodeRune(rb[:], up)
					rb = shiftNRuneBytes(rb, off)

					idxc := rb[0]
					for i, c := range []byte(n.indices) {
						// Uppercase matches
						if c == idxc {
							// Continue with child node
							n = n.children[i]
							npLen = len(n.path)
							continue walk
						}
					}
				}
			}

			// Nothing found. We can recommend to redirect to the same URL
			// without a trailing slash if a leaf exists for that path
			if fixTrailingSlash && path == "/" && n.handlers != nil {
				return ciPath
			}
			return nil
		}

		n = n.children[0]
		switch n.nType {
		case param:
			// Find param end (either '/' or path end)
			end := 0
			for end < len(path) && path[end] != '/' {
				end++
			}

			// Add param value to case insensitive path
			ciPath = append(ciPath, path[:end]...)

			// We need to go deeper!
			if end < len(path) {
				if len(n.children) > 0 {
					// Continue with child node
					n = n.children[0]
					npLen = len(n.path)
					path = path[end:]
					continue
				}

				// ... but we can't
				if fixTrailingSlash && len(path) == end+1 {
					return ciPath
				}
				return nil
			}

			if n.handlers != nil {
				return ciPath
			}

			if fixTrailingSlash && len(n.children) == 1 {
				// No handle found. Check if a handle for this path + a
				// trailing slash exists
				n = n.children[0]
				if n.path == "/" && n.handlers != nil {
					return append(ciPath, '/')
				}
			}

			return nil

		case catchAll:
			return append(ciPath, path...)

		default:
			panic("invalid node type")
		}
	}

	// Nothing found.
	// Try to fix the path by adding / removing a trailing slash
	if fixTrailingSlash {
		if path == "/" {
			return ciPath
		}
		if len(path)+1 == npLen && n.path[len(path)] == '/' &&
			strings.EqualFold(path[1:], n.path[1:len(path)]) && n.handlers != nil {
			return append(ciPath, n.path...)
		}
	}
	return nil
}
