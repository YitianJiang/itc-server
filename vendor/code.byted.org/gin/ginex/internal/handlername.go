package internal

import (
	"reflect"

	"github.com/gin-gonic/gin"
)

var (
	handlerNames    = make(map[uintptr]string)
	path2MethodDict = make(map[string]string)
)

func SetHandlerName(handler gin.HandlerFunc, name string) {
	handlerNames[reflect.ValueOf(handler).Pointer()] = name
}

func GetHandlerName(handler gin.HandlerFunc) string {
	return handlerNames[reflect.ValueOf(handler).Pointer()]
}

// 被Wrap修饰过的接口，函数地址都是相同的，会导致handlerNames不准确
// 新增Url绝对路径到handler Name的映射
func SetHandlerNameByPath(absolutePath string, name string) {
	path2MethodDict[absolutePath] = name
}

func GetHandlerNameByPath(absolutePath string) string {
	return path2MethodDict[absolutePath]
}
