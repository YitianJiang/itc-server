package lark

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

var errFuncDeprecated = errors.New("Function is deprecated")

func deprecateFunc(fn interface{}, msg string) error {
	funcName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	fmt.Printf("Warn: %s is deprecated.\n%s\n", funcName, msg)
	return errFuncDeprecated
}
