package ginex

import (
	"strconv"

	"code.byted.org/bagent/go-client"
	"code.byted.org/gin/ginex/internal"
	internal_util "code.byted.org/gin/ginex/internal/util"
	"code.byted.org/gopkg/logs"
)

func reportMetainfo(extra map[string]string) {
	infos := make(map[string]string)
	infos["psm"] = PSM()
	infos["cluster"] = internal_util.LocalCluster()
	infos["language"] = "go"
	infos["framework"] = "ginex"
	infos["framework_version"] = internal.VERSION
	infos["protocol"] = "http"
	infos["ip"] = LocalIP()
	infos["port"] = strconv.Itoa(appConfig.ServicePort)
	infos["debug_port"] = strconv.Itoa(appConfig.DebugPort)

	if extra != nil {
		for k, v := range extra {
			infos[k] = v
		}
	}

	defer func() {
		if r := recover(); r != nil {
			logs.Warn("Report server's metadata unsuccessfully, but it can be ignored: %v", r)
		}
	}()

	cli, err := bagentutil.NewClient()
	if err != nil {
		logs.Warnf("Failed to create bagent: %s", err)
		return
	}

	if err := cli.ReportInfo(infos); err != nil {
		logs.Warnf("Report server's metadata unsuccessfully, bu it can be ignored: %s", err)
	}
}
