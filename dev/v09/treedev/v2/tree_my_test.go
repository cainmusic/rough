package tree

import (
	"fmt"
	"testing"

	"github.com/cainmusic/gtable"
)

type debugNode struct {
	floor int
	node  *node
}

func (n *node) debugPrint() {
	l := &[]debugNode{}
	f := 0
	n.debugRead(l, f)
	debugTreePrint(l)
	//debugTreePrint2(l)
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
	table := gtable.NewTable()

	table.AppendHead([]string{"prio", "floor", "path", "fullPath", "handlers", "nType", "wildChild", "indices"})
	for _, d := range *l {
		table.AppendBody([]string{
			fmt.Sprint(d.node.priority),
			fmt.Sprint(d.floor),
			d.node.path,
			d.node.fullPath,
			fmt.Sprint(d.node.handlers),
			fmt.Sprint(d.node.nType),
			fmt.Sprint(d.node.wildChild),
			d.node.indices,
		})
	}

	table.PrintData()
}

/*
func debugTreePrint2(l *[]debugNode) {
	table := gtable.NewTable()

	tls := make([]gtable.TreeLayer, len(*l))
	for i, d := range *l {
		tls[i] = gtable.TreeLayer{Layer: d.floor, Name: d.node.path}
	}

	table.FormatTree(tls)

	table.SetNoBorder()

	table.PrintData()
}
*/

// 根catchAll
var pathCase00 = []string{
	"/*filepath",
}

// 根catchAll加兄弟
var pathCase01 = []string{
	"/*filepath",
	"/hello",
}

// 冲突，*前面没有自由的/
var pathCase0201 = []string{
	"/static/",
	"/static/*filepath",
}

var pathCase0202 = []string{
	"/static/*filepath",
	"/static/*otherpath",
}

// /static/*filepath相关
var pathCase0301 = []string{
	"/static",
	"/static/*filepath",
}

var pathCase0302 = []string{
	"/static/*filepath",
	"/static",
}

// param通配符1
var pathCase0401 = []string{
	"/user",
	"/user/:id",
}

// param通配符2
var pathCase0402 = []string{
	"/user/",
	"/user/:id",
}

// 特殊的param通配符
var pathCase05 = []string{
	":hello",
}

var pathCase06 = []string{
	"/get/:param/abc",
	"/get/:param",
}

func TestOne(t *testing.T) {
	n := new(node)

	for i, path := range test5 {
		fmt.Println(path)
		recv := catchPanic(func() { n.addRoute(path, i+1) })
		if recv != nil {
			fmt.Println("err", recv)
		}
		n.debugPrint()
	}
}

var test1 = []string{
	"/who/are/foo",
	"/who/are/foo/",
	"/who/are/foo/bar",
	"/con:nection",
}

var test2 = []string{
	"/files/:dir/*filepath",
}

var test3 = []string{
	"/vendor/:x/*y",
}

var test4 = []string{
	"/src1",
	"/src2*filepath",
	"/src2/*filepath",
}

var test5 = []string{
	"/hello/:ppaa",
	"/hello/pp",
}
