package discovery

import (
	"os"
	"strconv"
	"strings"

	"code.byted.org/golf/consul"
)

var (
	consulAgentHost string = "127.0.0.1"
	consulAgentPort int    = 2280
)

func init() {
	if host := os.Getenv("TCE_HOST_IP"); host != "" {
		consulAgentHost = host
	}

	if host := os.Getenv("CONSUL_HTTP_HOST"); host != "" {
		consulAgentHost = host
	}

	if strPort := os.Getenv("CONSUL_HTTP_PORT"); strPort != "" {
		port, err := strconv.Atoi(strPort)
		if err == nil {
			consulAgentPort = int(port)
		}
	}
}

// ConsulDiscover discover this service with specifical idc by consul
func ConsulDiscover(serviceName, idc string) ([]*Instance, error) {
	idc = strings.TrimSpace(idc)
	items, err := consul.TranslateOneOnHost(serviceName+".service."+idc, consulAgentHost, consulAgentPort)
	if err != nil {
		return nil, err
	}

	var ret []*Instance
	for _, ins := range items {
		ret = append(ret, NewInstance(ins.Host, strconv.Itoa(ins.Port), ins.Tags))
	}
	return ret, nil
}

// ConsulDiscoverer .
type ConsulDiscoverer struct{}

// NewConsulDiscoverer .
func NewConsulDiscoverer() *ConsulDiscoverer {
	return &ConsulDiscoverer{}
}

// Discover .
func (c *ConsulDiscoverer) Discover(serviceName, idc string) ([]*Instance, error) {
	return ConsulDiscover(serviceName, idc)
}
