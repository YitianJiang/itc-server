package discovery

import (
	"strconv"
	"strings"
)

// ServiceDiscoverer .
type ServiceDiscoverer interface {
	Discover(serviceName, idc string) ([]*Instance, error)
}

// Instance .
type Instance struct {
	Host string
	Port string
	Tags map[string]string
}

// NewInstance .
func NewInstance(host, port string, tags map[string]string) *Instance {
	for key, val := range tags {
		tags[key] = strings.TrimSpace(val)
	}
	return &Instance{
		Host: strings.TrimSpace(host),
		Port: strings.TrimSpace(port),
		Tags: tags,
	}
}

// Cluster return cluster name, if no cluster return "default"
func (it *Instance) Cluster() string {
	if it.Tags == nil {
		return "default"
	}
	cluster, ok := it.Tags["cluster"]
	if ok {
		if cluster == "" {
			return "default"
		}
		return cluster
	}
	return "default"
}

// Env return env name, if no env return "prod"
func (it *Instance) Env() string {
	if it.Tags == nil {
		return "prod"
	}
	env, ok := it.Tags["env"]
	if ok {
		if env == "" {
			env = "prod"
		}
		return env
	}
	return "prod"
}

// Weight return weight, default 100
func (it *Instance) Weight() int {
	const defaultWeight = 100
	if it.Tags == nil {
		return defaultWeight
	}

	if weight, ok := it.Tags["weight"]; ok {
		val, err := strconv.ParseInt(weight, 10, 64)
		if err != nil {
			val = defaultWeight
		}
		return int(val)
	}
	return defaultWeight
}
