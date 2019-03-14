package tccclient

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"code.byted.org/gopkg/env"
	etcdutil "code.byted.org/gopkg/etcd_util"
	etcd "code.byted.org/gopkg/etcd_util/client"
	"code.byted.org/gopkg/etcdproxy"
	"code.byted.org/gopkg/metrics"
)

const (
	ApiVersion       = "v1"
	Version          = "v1.0.2"
	TccEtcdKeyPrefix = "/tcc/" + ApiVersion + "/"

	metricsPrefix string = "tcc"
)

type Client struct {
	serviceName string
	cluster     string
	env         string

	cache *Cache

	disableMetrics bool
}

var (
	metricsClient *metrics.MetricsClient
	proxy         *etcdproxy.EtcdProxy
	needAgent     bool

	ConfigNotFoundError = errors.New("config not found error") // similar to Key Not Found error in etcd
	NetworkError        = errors.New("network error")

	clients     []*Client
	clientsLock sync.Mutex
)

func init() {
	// TODO: make sure etcd_util supports fallback from agent to etcd proxy
	// etcdutil.SetFallbackStrategy()

	metricsClient = metrics.NewDefaultMetricsClient(metricsPrefix, true)
	needAgent = checkNeedAgent()
	if !needAgent {
		proxy = etcdproxy.NewEtcdProxy()
	}
}

func checkNeedAgent() bool {
	confPath := "/etc/ss_conf"
	prodPath := "/opt/tiger/ss_conf/ss"

	p, err := os.Readlink(confPath)
	if err != nil {
		return true
	}

	if p != prodPath {
		return false
	}
	return true
}

type Listener func(key string, value string)

// NewClient returns tcc client
func NewClient(serviceName string, config *Config) (*Client, error) {
	client := Client{}
	client.serviceName = serviceName
	client.cluster = config.Cluster
	client.env = config.Env
	client.cache = NewCache()
	client.disableMetrics = config.DisableMetrics

	_, err := etcdutil.GetDefaultClient()
	if err != nil {
		return nil, err
	}

	checkClientConfig(&client)

	return &client, nil
}

func (c *Client) getRealKey(key string) string {
	return TccEtcdKeyPrefix + c.serviceName + "/" + c.cluster + "/" + key
}

func getDirectly(key string) (string, error) {
	value, err := proxy.Get(key)
	if err != nil {
		if etcdproxy.IsKeyNotFound(err) {
			return "", ConfigNotFoundError
		} else {
			return "", err
		}
	}
	return value, nil
}

// Get gets value by config key, may return error if the cnofig doesn't exist
func (c *Client) Get(key string) (string, error) {
	realKey := c.getRealKey(key)

	if c.env == "prod" && !needAgent {
		return getDirectly(realKey)
	}

	client, err := etcdutil.GetDefaultClient()
	if err != nil {
		return "", err
	}

	item := c.cache.Get(key)
	if item != nil {
		if !item.Expired() {
			return item.Value, nil
		}
	}

	ctx := context.Background()
	resp, err := client.Get(ctx, realKey, nil)
	if err != nil {
		if etcd.IsKeyNotFound(err) {
			c.emit(nil)
			return "", ConfigNotFoundError
		}
		c.emit(err)
		if item != nil {
			c.cache.Set(key, Item{Value: item.Value, Expires: time.Now().Add(5 * time.Second)})
			return item.Value, nil
		}
		return "", err
	}

	c.emit(nil)
	c.cache.Set(key, Item{Value: resp.Node.Value, Expires: time.Now().Add(5 * time.Second)})
	return resp.Node.Value, nil
}

func (c *Client) emit(err error) {

	if c.disableMetrics {
		return
	}

	result := "success"
	code := 0
	if err != nil {
		result = "failed"
		code = 1
	}

	key := "client.get_config." + result
	tagkv := map[string]string{
		"service_name": c.serviceName,
		"cluster":      c.cluster,
		"env":          c.env,
		"version":      Version,
		"api_version":  ApiVersion,
		"code":         fmt.Sprintf("%d", code),
		"language":     "go",
	}
	metricsClient.EmitCounter(key, 1, "", tagkv)
}

// Watch watches on local bagent
func (c *Client) Watch(key string, listener Listener) error {
	panic("Not Implemented")
}

// GetAllKeys return all config keys in this region
func (c *Client) GetAllKeys() []string {
	panic("Not Implemented")
}

// IsConfigNotFoundError returns whether an error is ConfigNotFoundError, which means the config
// has not been set
func IsConfigNotFoundError(err error) bool {
	return err == ConfigNotFoundError
}

// GetClusterFromEnv returns cluster reading from system env, "default" is returned
// when it's empty
func GetClusterFromEnv() string {
	return env.Cluster()
}

// GetServiceNameFromEnv returns service name reading from system env, "-" is returned
// when it's empty
func GetServiceNameFromEnv() string {
	return env.PSM()
}
