package conf

import (
	"code.byted.org/golf/ssconf"
	"fmt"
	"os"
)

type Config struct {
	MysqlConfigPath string
}

var (
	Configuration Config
)
//初始化配置信息
func InitConfiguration(){

	var err error
	conf, err := ssconf.LoadSsConfFile("conf/deploy.conf")
	if err != nil{
		fmt.Printf("load conf failed: %v\n", err)
	}
	if Configuration.MysqlConfigPath = conf["MysqlConfigPath"]; Configuration.MysqlConfigPath == ""{
		fmt.Printf("MysqlConfigPath not found in conf\n")
		os.Exit(-1)
	}
}