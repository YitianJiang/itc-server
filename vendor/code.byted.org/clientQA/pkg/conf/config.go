package conf

import (
	"code.byted.org/golf/ssconf"
	"fmt"
	"os"
	"runtime"
	"strconv"
)

type Config struct {
	Log           string
	Bings         []string
	MaxProc       int
	MysqlConfPath string
	ReloadTime    int
}

var (
	ProjectConfig Config
)

func InitConf() {
	var err error
	conf, err := ssconf.LoadSsConfFile("conf/deploy.conf")
	if err != nil {
		fmt.Printf("load conf failed: %v\n", err)
	}

	ProjectConfig.MaxProc = runtime.NumCPU()
	if maxProc, ok := conf["MaxProc"]; ok {
		if n, err := strconv.Atoi(maxProc); err == nil {
			ProjectConfig.MaxProc = n
		}
	}
	runtime.GOMAXPROCS(ProjectConfig.MaxProc)

	if ProjectConfig.MysqlConfPath = conf["MysqlConfPath"]; ProjectConfig.MysqlConfPath == "" {
		fmt.Printf("MySqlConfPath not found in conf\n")
		os.Exit(-1)
	}

	ProjectConfig.ReloadTime = 300
	if reloadTime, ok := conf["ReloadTime"]; ok {
		if n, err := strconv.Atoi(reloadTime); err == nil {
			ProjectConfig.ReloadTime = n
		}
	}

	ProjectConfig.Bings, err = ssconf.GetServersFromCache(conf, "binds")
	if err != nil || len(ProjectConfig.Bings) == 0 {
		fmt.Printf("binds not found in conf\n")
		os.Exit(-1)
	}
}
