package kitc

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"code.byted.org/gopkg/logs"
	"code.byted.org/kite/endpoint"
	"code.byted.org/kite/kitc/connpool"
	"code.byted.org/kite/kitc/discovery"
	"code.byted.org/kite/kitc/loadbalancer"
	"code.byted.org/kite/kitutil/cache"
	"code.byted.org/kite/kitutil/kiterrno"
)

var (
	enableFallback = true
)

func DisableFallback() {
	enableFallback = false
}

func makeFetchKey(service, idc, cluster, env string) string {
	return service + ":" + idc + ":" + cluster + ":" + env
}

// kitcDiscoverer 根据tag信息对discover发现的实例进行过滤, 并进行缓存
type kitcDiscoverer struct {
	kitcClient *KitcClient
	discover   discovery.ServiceDiscoverer
	cache      *cache.Asyncache
	policy     *discovery.DiscoveryPolicy
}

func newKitcDiscoverer(kclient *KitcClient, discover discovery.ServiceDiscoverer) *kitcDiscoverer {
	d := &kitcDiscoverer{
		discover:   discover,
		kitcClient: kclient,
	}
	c := cache.NewAsyncache(cache.Options{
		BlockIfFirst:    true,
		RefreshDuration: time.Second * 3,
		Fetcher:         d.fetch,
		ErrHandler:      d.fetchErrHandlerfunc,
		ChangeHandler:   d.instancesChangeHandler,
		IsSame:          d.isSameInstances,
	})
	d.cache = c
	return d
}

func (d *kitcDiscoverer) Discover(serviceName, idc, cluster, env string) ([]*discovery.Instance, error) {
	key := makeFetchKey(serviceName, idc, cluster, env)
	v := d.cache.Get(key, []*discovery.Instance{})
	ins := v.([]*discovery.Instance)

	if refresher, ok := d.kitcClient.loadbalancer.(loadbalancer.Rebalancer); ok {
		if !refresher.IsExist(key) {
			refresher.Rebalance(key, ins)
		}
	}

	if len(ins) == 0 {
		return nil, fmt.Errorf("no instance for service: %s, idc: %s, cluster: %s, env: %s",
			serviceName, idc, cluster, env)
	}

	copied := make([]*discovery.Instance, len(ins))
	copy(copied, ins)
	return copied, nil
}

func (d *kitcDiscoverer) Dump() map[string][]*discovery.Instance {
	data := d.cache.Dump()
	result := make(map[string][]*discovery.Instance)
	for k, v := range data {
		result[k] = v.([]*discovery.Instance)
	}
	return result
}

func (d *kitcDiscoverer) fetchErrHandlerfunc(key string, err error) {
	if strings.Contains(key, ":hl:") {
		logs.Warnf("KITC: service discover key: %s, err: %s", key, err.Error())
		return
	}
	logs.Errorf("KITC: service discover key: %s, err: %s", key, err.Error())
}

func (d *kitcDiscoverer) isSameInstances(key string, oldData, newData interface{}) bool {
	oldIns := oldData.([]*discovery.Instance)
	newIns := newData.([]*discovery.Instance)
	if len(oldIns) != len(newIns) {
		return false
	}
	if len(oldIns) == 0 && len(newIns) == 0 {
		return true
	}

	oldStrs := make([]string, 0, len(oldIns))
	for _, i := range oldIns {
		oldStrs = append(oldStrs, i.Host+":"+i.Port+":"+string(i.Weight()))
	}
	newStrs := make([]string, 0, len(newIns))
	for _, i := range newIns {
		newStrs = append(newStrs, i.Host+":"+i.Port+":"+string(i.Weight()))
	}

	sort.Strings(oldStrs)
	sort.Strings(newStrs)

	for i := range oldStrs {
		if oldStrs[i] != newStrs[i] {
			return false
		}
	}

	return true
}

func (d *kitcDiscoverer) instancesChangeHandler(key string, oldData, newData interface{}) {
	oldIns := oldData.([]*discovery.Instance)
	newIns := newData.([]*discovery.Instance)
	d.kitcClient.discoverChangeHandler(key, oldIns, newIns)

	if refresher, ok := d.kitcClient.loadbalancer.(loadbalancer.Rebalancer); ok {
		refresher.Rebalance(key, newIns)
	}

	if longPool, ok := d.kitcClient.pool.(connpool.LongConnPool); ok {
		insMap := make(map[string]struct{}, len(newIns))
		for _, in := range newIns {
			insMap[in.Host+":"+in.Port] = struct{}{}
		}

		for _, in := range oldIns {
			addr := in.Host + ":" + in.Port
			if _, ok := insMap[addr]; !ok {
				longPool.Clean(in.Host, in.Port)
			}
		}
	}
}

func (d *kitcDiscoverer) fetch(key string) (interface{}, error) {
	tmp := strings.Split(key, ":")
	if len(tmp) != 4 {
		return nil, fmt.Errorf("KITC: invalid key when discover: %s", key)
	}
	service, idc, cluster, env := tmp[0], tmp[1], tmp[2], tmp[3]

	ins, err := d.discover.Discover(service, idc)
	if err != nil {
		logs.Warnf("KITC: discover call failed, err:%v", err)
		return nil, err
	}
	if len(ins) == 0 {
		// hard-code fallback
		if enableFallback && idc == "hl" {
			logs.Warnf("KITC: service discovery: no instance in hl, fallback to lf (hard-coded), service:%s", key)
			return d.Discover(service, "lf", cluster, env)
		}

		return nil, fmt.Errorf("KITC: no instance on discover, service:%s", key)
	}

	filtered := d.policy.Filter(ins, cluster, env)
	if len(filtered) == 0 {
		return nil, fmt.Errorf("KITC: no instance remains for %s, ins list: %s", key, instances2ReadableDetail(ins))
	}

	return filtered, nil
}

func instances2ReadableDetail(ins []*discovery.Instance) string {
	if len(ins) == 0 {
		return "[]"
	}
	insStr := make([]string, 0, len(ins))
	for _, i := range ins {
		insStr = append(insStr, i.Host+":"+i.Port+fmt.Sprintf("{%v}", i.Tags))
	}
	return strings.Join(insStr, ",")
}

func instances2ReadableStr(ins []*discovery.Instance) string {
	if len(ins) == 0 {
		return "[]"
	}
	insStr := make([]string, 0, len(ins))
	for _, i := range ins {
		insStr = append(insStr, i.Host+":"+i.Port)
	}
	sort.Strings(insStr)
	return strings.Join(insStr, ",")
}

// NewServiceDiscoverMW .
func NewServiceDiscoverMW(discoverer *kitcDiscoverer) endpoint.Middleware {
	return func(next endpoint.EndPoint) endpoint.EndPoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			rpcInfo := GetRPCInfo(ctx)
			if len(rpcInfo.Instances) > 0 {
				return next(ctx, request)
			}

			ins, err := discoverer.Discover(rpcInfo.To, rpcInfo.TargetIDC, rpcInfo.ToCluster, rpcInfo.Env)
			if err != nil {
				kerr := kiterrno.NewKitErr(kiterrno.ServiceDiscoverCode,
					fmt.Errorf("service discover idc=%s service=%s cluster=%s env=%s err: %s",
						rpcInfo.TargetIDC, rpcInfo.To, rpcInfo.ToCluster, rpcInfo.Env, err.Error()))
				return kiterrno.ErrRespServiceDiscover, kerr
			}
			if len(ins) == 0 {
				kerr := kiterrno.NewKitErr(kiterrno.ServiceDiscoverCode,
					fmt.Errorf("No service discovered idc=%s service=%s cluster=%s env=%s",
						rpcInfo.TargetIDC, rpcInfo.To, rpcInfo.ToCluster, rpcInfo.Env))
				return kiterrno.ErrRespServiceDiscover, kerr
			}

			rpcInfo.Instances = ins
			return next(ctx, request)
		}
	}
}
