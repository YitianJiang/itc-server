package tools

import (
	"code.byted.org/gopkg/consul"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/pkg/errors"
	"context"
	"math/rand"
)

type TargetService struct {
	HostPort string
	Cluster  string
}

func DiscoverAllServices(ctx context.Context, psm string) []TargetService {
	var retServices []TargetService

	endpoints, err := consul.Lookup(psm)
	if err != nil {
		logs.CtxError(ctx, "Consul Lookup error:%v, psm=%s", err, psm)
		return retServices
	}
	for _, endpoint := range endpoints {
		retServices = append(retServices, TargetService{
			HostPort: endpoint.Addr,
			Cluster:  endpoint.Cluster,
		})
	}
	return retServices
}

//从consul获取psm的服务列表，随机选择一个返回
func GetServerAddr(ctx context.Context, psm string) (string, error) {
	allServices := DiscoverAllServices(ctx, psm)
	servicesNum := len(allServices)
	var serviceAddr string
	if servicesNum > 1 {
		serviceAddr = allServices[rand.Intn(servicesNum)].HostPort
	} else if servicesNum == 1 {
		serviceAddr = allServices[0].HostPort
	}
	if servicesNum == 0 {
		logs.CtxError(ctx, "Consul Lookup empty")
		return "", errors.New("Consul Lookup empty")
	}
	return serviceAddr, nil
}
