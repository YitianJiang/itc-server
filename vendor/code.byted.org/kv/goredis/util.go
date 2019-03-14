package goredis

import (
	"context"
	"fmt"
	"os"
	"strings"

	redis "code.byted.org/kv/redis-v6"
)

const (
	IDC_URL = "http://get-idc.d.byted.org/map/json"

	STRESSTAG = "K_STRESS"
)

type IdcInfo struct {
	County   string `json:"county"`
	Provider string `json:"privode"`
	// Provider string `json:"provider"`
	City string `json:"city"`
	Name string `json:"name"`
}

var (
	isStressTestEnv       = os.Getenv("TCE_PERF_TEST")
	stressTestPrefix      = os.Getenv("TCE_PERF_PREFIX")
	stressTestWhiteList   = os.Getenv("TCE_PERF_WHITELIST")
	isStressTestWhiteList = false
)

func getStressTag(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}

	v, ok := ctx.Value(STRESSTAG).(string)
	if !ok {
		return "", false
	}
	return v, ok
}

func isInWhiteList(cluster string) {
	for _, part := range strings.Split(stressTestWhiteList, "&") {
		if part == fmt.Sprintf("cache:%s", cluster) {
			isStressTestWhiteList = true
			return
		}
	}
	isStressTestWhiteList = false
	return
}

func isStressTest(ctx context.Context) (string, bool) {
	if isStressTestEnv != "" && isStressTestWhiteList == false {
		return stressTestPrefix, true
	}

	return getStressTag(ctx)
}

func convertStressCMD(prefix string, cmd redis.Cmder) redis.Cmder {
	args := cmd.Args()
	if len(args) < 2 {
		return cmd
	}
	if !strings.HasSuffix(prefix, "_") {
		prefix = prefix + "_"
	}

	method := strings.ToLower(cmd.Name())
	if method == "del" || method == "mget" {
		for i := 1; i < len(args); i++ {
			key := args[i].(string)
			args[i] = prefix + key
		}
	} else {
		key := args[1].(string)
		args[1] = prefix + key
	}

	return cmd
}
