package goredis

import (
	"errors"
	"strings"
	"time"

	"code.byted.org/golf/ssconf"
	"code.byted.org/gopkg/logs"

	"code.byted.org/kv/redis-v6"
)

const (
	REDIS_WEB_CONF_PATH = "/opt/tiger/ss_conf/ss/redis_web.conf"
	REDIS_CONF_PATH     = "/opt/tiger/ss_conf/ss/redis.conf"
	REDIS_CONF          = "redis_conf"

	IDC_KEYWORD = ".service."
)

func autoLoadConf(cluster string, ch chan []string, option *Option) {
	if cluster == "" {
		logs.Error(ErrEmptyClusterName.Error())
		return
	}
	go func() {
		ticker := time.NewTicker(option.autoLoadInterval)
		defer func() {
			ticker.Stop()
		}()
		for {
			select {
			case <-ticker.C:
				servers, err := loadConfByClusterName(cluster, option.configFilePath, option.useConsul)
				if err != nil {
					logs.Errorf("redis client auto load conf error, cluster:%s, err:%v", cluster, err)
				} else if len(servers) > 2 {
					//logs.Infof("autoLoadConf, cluster:%v, servers:%v\n", cluster, servers)
					ch <- servers
				}
			}
		}
	}()
}

func loadConfByClusterName(clusterName, userDefinedConfPath string, useConsul bool) (servers []string, err error) {
	if len(clusterName) == 0 {
		return nil, ErrEmptyClusterName
	}
	consulClusterName := checkConsulClusterName(clusterName)

	// get from ssconf if useConsul is false
	if !useConsul {
		if len(userDefinedConfPath) > 0 {
			servers, err = ssconf.GetServerList(userDefinedConfPath, clusterName)
		}
		if len(servers) == 0 || err != nil {
			servers, err = ssconf.GetServerList(REDIS_WEB_CONF_PATH, clusterName)
		}
		if len(servers) == 0 || err != nil {
			servers, err = ssconf.GetServerList(REDIS_CONF_PATH, clusterName)
		}
	}

	// get from consul
	if len(servers) == 0 || err != nil {
		servers, err = serviceDiscoveryByConsul(consulClusterName)
	}

	if err != nil {
		return nil, err
	}
	if len(servers) == 0 {
		return nil, ErrClusterConfigNotFound
	}
	return servers, nil
}

func serviceDiscoveryByConsul(clusterName string) ([]string, error) {
	// get idc from clusterName: cluster.service.idc
	idc := LocalIDC()
	if idx := strings.Index(clusterName, IDC_KEYWORD); idx != -1 {
		idc = clusterName[idx+len(IDC_KEYWORD):]
		clusterName = clusterName[:idx]
	}

	consulAgent := NewConsulService(clusterName)
	// prefer local idc
	var insts []*Instance = nil
	insts = consulAgent.Lookup(idc)
	if insts == nil {
		return nil, ErrConsulServerEmpty
	}
	var servers []string
	for _, inst := range insts {
		servers = append(servers, inst.Str())
	}
	return servers, nil
}

func checkConsulClusterName(clusterName string) string {
	clusterNameInPSM := ""
	if strings.HasPrefix(clusterName, REDIS_PREFIX) {
		clusterNameInPSM = REDIS_PSM_PREFIX + clusterName[len(REDIS_PREFIX):]
	} else if strings.HasPrefix(clusterName, REDIS_PSM_PREFIX) {
		clusterNameInPSM = clusterName
	} else { // for abase tmp
		clusterNameInPSM = clusterName
	}
	return clusterNameInPSM
}
func serviceDiscoveryByRedis(clusterName string) ([]string, error) {
	servers, err := ssconf.GetServerList(REDIS_WEB_CONF_PATH, REDIS_CONF)
	if len(servers) == 0 || err != nil {
	}
	if err != nil {
		return nil, err
	}
	if len(servers) == 0 {
		return nil, errors.New("[serviceDiscoveryByRedis] get servers for ss_conf err")
	}
	client := redis.NewClient(&redis.Options{
		Addr: servers[0],
	})
	cmd := redis.NewStringCmd("HGET", clusterName, "proxys")
	client.Process(cmd)
	v, err := cmd.Result()
	return strings.Split(v, ","), err
}
