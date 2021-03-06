package kitc

import (
	"encoding/json"
	"fmt"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"code.byted.org/gopkg/env"
	"code.byted.org/gopkg/logs"
	"code.byted.org/kite/kitutil/cache"
	"code.byted.org/kite/kitutil/kitevent"
	"code.byted.org/kite/kitutil/kvstore"
)

var defaultMeshProxyConfig = RPCConfig{
	RPCTimeout:      120000,
	ConnectTimeout:  30000,
	ReadTimeout:     60000,
	WriteTimeout:    60000,
	StressBotSwitch: false,
}

var defaultRPCConfig = RPCConfig{
	RPCTimeout:     1000,
	ConnectTimeout: 50,
	ReadTimeout:    1000,
	WriteTimeout:   1000,
	IDCConfig: []IDCConfig{IDCConfig{
		IDC:     env.IDC(),
		Percent: 100,
	}},
	ServiceCB: CBConfig{
		IsOpen:         true,
		ErrRate:        0.5,
		MinSample:      200,
		MaxConcurrency: 1000,
	},
	ACLAllow:        true,
	DegraPercent:    0,
	StressBotSwitch: false,
}

// remoteConfiger manage all remote configs
type remoteConfiger struct {
	kitcClient *KitcClient
	kvstorer   kvstore.KVStorer
	cache      *cache.Asyncache
}

func newRemoteConfiger(kclient *KitcClient, kvstorer kvstore.KVStorer) *remoteConfiger {
	c := &remoteConfiger{
		kvstorer:   kvstorer,
		kitcClient: kclient,
	}
	c.cache = cache.NewAsyncache(cache.Options{
		BlockIfFirst:    true,
		RefreshDuration: time.Second * 10,
		Fetcher:         c.fetchRemoteConfig,
		ErrHandler:      c.getRemoteErrHandler,
		ChangeHandler:   c.remoteConfigChangeHandler,
		IsSame:          c.isSameRemoteConfig,
	})
	return c
}

func (kc *remoteConfiger) GetRemoteConfig(r RPCMeta) (RPCConfig, error) {
	key := r.ConfigKey()
	v := kc.cache.Get(key, defaultRPCConfig)
	return v.(RPCConfig), nil
}

// GetAllRemoteConfigs .
func (kc *remoteConfiger) GetAllRemoteConfigs() map[string]RPCConfig {
	data := kc.cache.Dump()
	configs := make(map[string]RPCConfig)
	for k, v := range data {
		configs[k] = v.(RPCConfig)
	}
	return configs
}

// getRPCConfig get remote config concurrencily
func (kc *remoteConfiger) fetchRemoteConfig(key string) (interface{}, error) {
	// get config from remote(ETCD)
	tmp := strings.Split(key, "/")
	if len(tmp) != 5 {
		return nil, fmt.Errorf("invalid RPC config key: %s", key)
	}
	r := RPCMeta{
		From:        tmp[0],
		FromCluster: tmp[1],
		To:          tmp[2],
		ToCluster:   tmp[3],
		Method:      tmp[4],
	}

	c := RPCConfig{}
	var wg sync.WaitGroup
	var fetchErr error
	var lock sync.Mutex
	wg.Add(1)
	go func() {
		defer wg.Done()
		tmIDC, err := kc.getTimeoutsAndIDCConf(r)
		if err != nil {
			lock.Lock()
			defer lock.Unlock()
			fetchErr = fmt.Errorf("KITC: get timeouts and IDC configs err: %s", err.Error())
			return
		}
		c.ConnectTimeout = tmIDC.ConnectTimeout
		c.WriteTimeout = tmIDC.WriteTimeout
		c.ReadTimeout = tmIDC.ReadTimeout
		c.RPCTimeout = tmIDC.WriteTimeout
		if tmIDC.ReadTimeout > c.RPCTimeout {
			c.RPCTimeout = tmIDC.ReadTimeout
		}
		c.IDCConfig = tmIDC.TrafficPolicy
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		cbSwitch, err := kc.getServiceCBSwitch(r)
		if err != nil {
			lock.Lock()
			defer lock.Unlock()
			fetchErr = fmt.Errorf("KITC: get circuitbreaker switch err: %s", err.Error())
			return
		}
		c.ServiceCB.IsOpen = cbSwitch
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		cbErrRate, err := kc.getServiceCBErrRate(r)
		if err != nil {
			lock.Lock()
			defer lock.Unlock()
			fetchErr = fmt.Errorf("KITC: get circuitbreaker err rate err: %s", err.Error())
			return
		}
		c.ServiceCB.ErrRate = cbErrRate
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		cbMinSample, err := kc.getServiceCBMinSample(r)
		if err != nil {
			lock.Lock()
			defer lock.Unlock()
			fetchErr = fmt.Errorf("KITC: get circuitbreaker min sample err: %s", err.Error())
			return
		}
		lock.Lock()
		defer lock.Unlock()
		c.ServiceCB.MinSample = cbMinSample
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		acl, err := kc.getACL(r)
		if err != nil {
			lock.Lock()
			defer lock.Unlock()
			fetchErr = fmt.Errorf("KITC: get acl err: %s", err.Error())
			return
		}
		lock.Lock()
		defer lock.Unlock()
		c.ACLAllow = acl
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		per, err := kc.getDegraPercent(r)
		if err != nil {
			lock.Lock()
			defer lock.Unlock()
			fetchErr = fmt.Errorf("KITC: get degradation percent err: %s", err.Error())
			return
		}
		lock.Lock()
		defer lock.Unlock()
		c.DegraPercent = per
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		stress, err := kc.getStressBotSwitch(r)
		if err != nil {
			lock.Lock()
			defer lock.Unlock()
			fetchErr = fmt.Errorf("KITC: get stress bot switch err: %s", err.Error())
			return
		}
		lock.Lock()
		defer lock.Unlock()
		c.StressBotSwitch = stress
	}()

	wg.Wait()
	return c, fetchErr
}

func (kc *remoteConfiger) getRemoteErrHandler(key string, err error) {
	if strings.Contains(err.Error(), "key not found") {
		return
	}

	logs.Warnf("KITC: fetch remote config key: %s, err: %s", key, err.Error())
}

func (kc *remoteConfiger) remoteConfigChangeHandler(key string, oldData, newData interface{}) {
	ocbuf, _ := json.Marshal(oldData)
	ncbuf, _ := json.Marshal(newData)
	e := &kitevent.KitEvent{
		Name:   "remote_config_change",
		Time:   time.Now(),
		Detail: fmt.Sprintf("%s: %s -> %s", key, string(ocbuf), string(ncbuf)),
	}
	kc.kitcClient.pushEvent(e)
}

func (kc *remoteConfiger) isSameRemoteConfig(key string, oldData, newData interface{}) bool {
	oc := oldData.(RPCConfig)
	nc := newData.(RPCConfig)
	if oc.RPCTimeout != nc.RPCTimeout {
		return false
	}
	if oc.ConnectTimeout != nc.ConnectTimeout {
		return false
	}
	if oc.ReadTimeout != nc.ReadTimeout {
		return false
	}
	if oc.WriteTimeout != nc.WriteTimeout {
		return false
	}
	if oc.ServiceCB != nc.ServiceCB {
		return false
	}
	if oc.ACLAllow != nc.ACLAllow {
		return false
	}
	if oc.DegraPercent != nc.DegraPercent {
		return false
	}

	if len(oc.IDCConfig) != len(nc.IDCConfig) {
		return false
	}
	if len(oc.IDCConfig) == 0 {
		return true
	}
	idcConfMap := make(map[string]int, len(oc.IDCConfig))
	for _, c := range oc.IDCConfig {
		idcConfMap[c.IDC] = c.Percent
	}
	for _, c := range nc.IDCConfig {
		v, ok := idcConfMap[c.IDC]
		if !ok {
			return false
		}
		if v != c.Percent {
			return false
		}
	}

	return true
}

// TODO(zhangyuanjia): ???????????????????????????, ?????????????????????, ??????????????????;
// ????????????, ????????????????????????????????????, ??????????????????????????????;
type tmAndIDC struct {
	RetryTimes          int // counts of retry
	ConnectTimeout      int // ms
	ConnectRetryMaxTime int // ms
	ReadTimeout         int // ms
	WriteTimeout        int // ms
	TrafficPolicy       []IDCConfig
}

var (
	defaultTMAndIDC    tmAndIDC
	defaultTMAndIDCStr string
)

func init() {
	defaultTMAndIDC = tmAndIDC{
		RetryTimes:          0,    // deprecated
		ConnectRetryMaxTime: 1000, // deprecated
		ConnectTimeout:      defaultRPCConfig.ConnectTimeout,
		ReadTimeout:         defaultRPCConfig.ReadTimeout,
		WriteTimeout:        defaultRPCConfig.WriteTimeout,
		TrafficPolicy:       defaultRPCConfig.IDCConfig,
	}

	buf, _ := json.Marshal(defaultTMAndIDC)
	defaultTMAndIDCStr = string(buf)
}

func (kc *remoteConfiger) getStressBotSwitch(r RPCMeta) (bool, error) {
	global := "/kite/stressbot/request/switch/global"
	val, globalErr := kc.kvstorer.GetOrCreate(global, "off")
	if globalErr != nil {
		return false, globalErr
	}
	if val == "off" {
		return false, nil
	}
	if val != "on" {
		return false, fmt.Errorf("invalid global stress switch value: %v", val)
	}

	psmKey := fmt.Sprintf("/kite/stressbot/%s/%s/request/switch", r.To, r.ToCluster)
	val, psmErr := kc.kvstorer.GetOrCreate(psmKey, "off")
	if psmErr != nil {
		return false, psmErr
	}
	if val == "off" {
		return false, nil
	} else if val == "on" {
		return true, nil
	}
	return false, fmt.Errorf("invalid psm stress switch value: %v", val)
}

func (kc *remoteConfiger) getTimeoutsAndIDCConf(r RPCMeta) (tmAndIDC, error) {
	key := confETCDPath(r)
	key = path.Join("/kite/config", env.IDC(), key)
	val, err := kc.kvstorer.GetOrCreate(key, defaultTMAndIDCStr)
	if err != nil {
		return tmAndIDC{}, err
	}

	var conf tmAndIDC
	if err := json.Unmarshal([]byte(val), &conf); err != nil {
		return tmAndIDC{}, fmt.Errorf("invalid etcd value: %s", val)
	}

	if conf.ReadTimeout <= 0 {
		logs.Errorf("KITC: invalid read timeout=%v, use default value %vms, to=%v, to_cluster=%v, method=%v", conf.ReadTimeout, defaultRPCConfig.ReadTimeout, r.To, r.ToCluster, r.Method)
		conf.ReadTimeout = defaultRPCConfig.ReadTimeout
	}
	if conf.WriteTimeout <= 0 {
		logs.Errorf("KITC: invalid write timeout=%v, use default value %vms, to=%v, to_cluster=%v, method=%v", conf.WriteTimeout, defaultRPCConfig.WriteTimeout, r.To, r.ToCluster, r.Method)
		conf.WriteTimeout = defaultRPCConfig.WriteTimeout
	}
	if conf.ConnectTimeout <= 0 {
		logs.Errorf("KITC: invalid connection timeout=%v, use default value %vms, to=%v, to_cluster=%v, method=%v", conf.ConnectTimeout, defaultRPCConfig.ConnectTimeout, r.To, r.ToCluster, r.Method)
		conf.ConnectTimeout = defaultRPCConfig.ConnectTimeout
	}

	// if conf.TrafficPolicy is nil, selectIDC will use local IDC
	return conf, nil
}

func (kc *remoteConfiger) getServiceCBSwitch(r RPCMeta) (bool, error) {
	key := path.Join("/kite/circuitbreaker/switch", confETCDPath(r))
	val, err := kc.kvstorer.GetOrCreate(key, "1")
	if err != nil {
		return false, err
	}

	if val == "1" {
		return true, nil
	} else if val == "0" {
		return false, nil
	}

	return false, fmt.Errorf("invalid circuitbreaker switch value: %s", val)
}

func (kc *remoteConfiger) getServiceCBErrRate(r RPCMeta) (float64, error) {
	key := path.Join("/kite/circuitbreaker/config", confETCDPath(r), "errRate")
	val, err := kc.kvstorer.GetOrCreate(key, "0.5")
	if err != nil {
		return 0, err
	}
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid circuitbreaker error rate value: %s", val)
	}

	return f, nil
}

func (kc *remoteConfiger) getServiceCBMinSample(r RPCMeta) (int, error) {
	key := path.Join("/kite/circuitbreaker/config", confETCDPath(r), "minSample")
	val, err := kc.kvstorer.GetOrCreate(key, "200")
	if err != nil {
		return 0, err
	}
	num, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid circuitbreaker min sample value: %s", val)
	}

	return num, nil
}

func (kc *remoteConfiger) getACL(r RPCMeta) (bool, error) {
	key := path.Join("/kite/acl", confETCDPath(r))
	val, err := kc.kvstorer.GetOrCreate(key, "0")
	if err != nil {
		return false, err
	}

	if val == "0" {
		return true, nil
	} else if val == "1" {
		return false, nil
	}
	return false, fmt.Errorf("invalid acl value: %s", val)
}

func (kc *remoteConfiger) getDegraPercent(r RPCMeta) (int, error) {
	key := path.Join("/kite/switches", confETCDPath(r))
	val, err := kc.kvstorer.GetOrCreate(key, "0")
	if err != nil {
		return 0, err
	}

	per, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid degradation percent value: %s", val)
	}
	return per, err
}

func confETCDPath(r RPCMeta) string {
	buf := make([]byte, 0, 100)
	buf = append(buf, r.From...)
	buf = append(buf, '/')
	if r.FromCluster != "default" && r.FromCluster != "" {
		buf = append(buf, r.FromCluster...)
		buf = append(buf, '/')
	}
	buf = append(buf, r.To...)
	buf = append(buf, '/')
	if r.ToCluster != "default" && r.ToCluster != "" {
		buf = append(buf, r.ToCluster...)
		buf = append(buf, '/')
	}
	buf = append(buf, r.Method...)
	return string(buf)
}
