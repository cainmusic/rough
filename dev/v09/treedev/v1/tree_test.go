package tree

import (
	"fmt"
	"testing"
)

var rootN *node = new(node)

var pathList = []string{
	"/hello",
	"/see",
	"/search",
	"/se",
	"/see/me",
	"/hello/world",
}

func TestTree(t *testing.T) {
	for i, path := range pathList {
		ti := i + 1
		rootN.addRoute(path, []int{ti})
		rootN.debugPrint()
		fmt.Println()
	}

	t.Logf(fmt.Sprintln(rootN.getRoute("/see")))
}
