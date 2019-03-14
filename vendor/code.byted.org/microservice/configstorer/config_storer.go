package configstorer

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"code.byted.org/gopkg/asyncache"
	"code.byted.org/gopkg/etcd_util"
)

var (
	configCache *asyncache.SingleAsyncCache
	mutex       sync.Mutex
)

const (
	keyNotFound = "__cs_KEY_NOT_FOUND__"
)

func configGetter(key string) (interface{}, error) {
	cli, err := etcdutil.GetDefaultClient()
	if err != nil {
		return "", err
	}

	node, err := cli.Get(context.Background(), key, nil)

	if err != nil {
		if isEtcdKeyNonexist(err) {
			return keyNotFound, nil
		}
		return "", err
	}
	return node.Node.Value, nil
}

func InitStorer() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Create config storer of etcd error. %v", r)
		}
	}()

	if configCache == nil {
		mutex.Lock()
		defer mutex.Unlock()
		if configCache == nil {
			configCache = asyncache.NewSingleAsyncCache(configGetter)
			etcdutil.SetRequestTimeout(50 * time.Millisecond)
		}
	}

	err = nil
	return
}

func Get(key string) (string, error) {
	val, err := configCache.Get(key)
	if val == nil {
		return "", err
	}
	if vstr, ok := val.(string); ok {
		if vstr == keyNotFound {
			return "", asyncache.EmptyErr
		}
		return vstr, err
	} else {
		return "", errors.New("value not string")
	}
}

func GetOrDefault(key, defaultVal string) (string, error) {
	val, err := Get(key)
	if IsKeyNonexist(err) {
		return defaultVal, nil
	}
	return val, err
}

func isEtcdKeyNonexist(err error) bool {
	return strings.Index(err.Error(), "100:") == 0
}

func IsKeyNonexist(err error) bool {
	return err == asyncache.EmptyErr
}

func SetGetterTimeout(duration time.Duration) {
	etcdutil.SetRequestTimeout(duration)
}
