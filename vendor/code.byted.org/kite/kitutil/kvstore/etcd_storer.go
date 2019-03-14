package kvstore

import (
	"context"
	"fmt"

	etcdutil "code.byted.org/gopkg/etcd_util"
	etcdclient "code.byted.org/gopkg/etcd_util/client"
)

type etcdStorer struct{}

// NewETCDStorer .
func NewETCDStorer() KVStorer {
	return &etcdStorer{}
}

func (e *etcdStorer) Get(key string) (string, error) {
	cli, err := etcdutil.GetDefaultClient()
	if err != nil {
		return "", err
	}
	val, err := cli.Get(context.Background(), key, nil)
	if err != nil {
		return "", err
	}
	return val.Node.Value, nil
}

func (e *etcdStorer) GetOrCreate(key, val string) (string, error) {
	v, err := e.Get(key)
	if err != nil && etcdclient.IsKeyNotFound(err) {
		// key not found is not defined as an error
		// don't create KV
		return val, nil
	}

	if err != nil {
		return "", fmt.Errorf("get key=%s err: %s", key, err.Error())
	}

	return v, nil
}
