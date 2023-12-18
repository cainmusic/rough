package rough

import (
	"path"
	"reflect"
	"runtime"
	"unsafe"
)

func lastChar(str string) uint8 {
	if str == "" {
		panic("The length of the string can't be 0")
	}
	return str[len(str)-1]
}

func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	if lastChar(relativePath) == '/' && lastChar(finalPath) != '/' {
		return finalPath + "/"
	}
	return finalPath
}

func StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

func BytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// 使用gin.H快速声明一个map
type H map[string]any

// 获取一个engine级别的key
const EnKey = "__rough_engine"

func GetEnKey(key string) string {
	return EnKey + "." + key
}

// assert
func assert1(guard bool, text string) {
	if !guard {
		panic(text)
	}
}

func nameOfFunction(f any) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
