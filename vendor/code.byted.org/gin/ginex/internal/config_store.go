package internal

import (
	"fmt"
	"time"

	"code.byted.org/microservice/configstorer"
)

func InitConfigStorer() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Create config storer of etcd error. %v", r)
		}
	}()

	configstorer.InitStorer()
	configstorer.SetGetterTimeout(25 * time.Millisecond)

	err = nil
	return
}
