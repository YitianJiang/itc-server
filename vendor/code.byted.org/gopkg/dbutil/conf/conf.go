package conf

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"code.byted.org/golf/ssconf"
	"code.byted.org/gopkg/logs"
	"gopkg.in/yaml.v2"
)

type DBOptional struct {
	DriverName   string `yaml:"DriverName"`
	Timeout      string `yaml:"Timeout"`      //Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	ReadTimeout  string `yaml:"ReadTimeout"`  //Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	WriteTimeout string `yaml:"WriteTimeout"` //Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
	User         string `yaml:"User"`
	Password     string `yaml:"Password"`
	DBName       string `yaml:"DBName"`
	DBCharset    string `yaml:"DBCharset"`
	DBHostname   string `yaml:"DBHostname"`
	DBPort       string `yaml:"DBPort"`
	MaxIdleConns int    `yaml:"MaxIdleConns"`
	MaxOpenConns int    `yaml:"MaxOpenConns"`
	ExtraDSNMap  map[string]string
}

func GetDefaultDBOptional() DBOptional {
	return DBOptional{
		DriverName:   "mysql",
		Timeout:      "100ms",
		ReadTimeout:  "2.0s",
		WriteTimeout: "5.0s",
		DBHostname:   "localhost",
		DBPort:       "3306",
		DBCharset:    "utf8", // use utf8 as default
		MaxIdleConns: 10,
		MaxOpenConns: 100,
	}
}

/**
* 初始化通过数据库授权的服务 mysqlConsul 数据库consul信息 toutiao.mysql.****_write,  servicePsm 本机服务信息，  authKey 本季服务拥有的对数据库服务的访问密钥匙
  可选 GetDBOptionalByConsul("toutiao.mysql.****_write","","") TCE 环境  servicePsm 可以为空 ，authKey 可以为空
       GetDBOptionalByConsul("toutiao.mysql.****_write","p.s.m","********") 线下测试 其他环境跑 都不为空
*/
func GetDBOptionalByConsul(mysqlConsul string, servicePsm string, authKey string) (DBOptional, error) {
	opt := GetDefaultDBOptional()
	opt.DBHostname = mysqlConsul
	opt.User = servicePsm
	opt.Password = authKey
	opt.DBName = ""
	opt.DBPort = ""
	opt.DriverName = "mysql2"
	return opt, nil
}

/**
* 初始化通过数据库授权的服务 mysqlConsul 数据库consul信息 toutiao.mysql.****_write,这个接口测试上线需要把 TCE_PSM/TOUTIAO_MYSQL_*****WRITE_AUTHKEY 注入环境变量
   GetDBOptionalByConsul("toutiao.mysql.****_write") TCE 环境
*/

func GetDBOptionalByConsulName(mysqlConsul string) (DBOptional, error) {
	if mysqlConsul == "" || !isConsulName(mysqlConsul) {
		return DBOptional{}, fmt.Errorf("GetDBOptionalByConsul Unsupport mysqlConsul [%s],expect toutiao.mysql.***_write/read/offline", mysqlConsul)
	}

	opt := GetDefaultDBOptional()
	opt.DBHostname = mysqlConsul
	opt.User = ""
	opt.Password = ""
	opt.DBName = ""
	opt.DBPort = ""
	opt.DriverName = "mysql2"
	return opt, nil
}

/**
 * 构造访问数据库配置，schema：[user[:password]@][net[(addr)]]/dbname[?param1=value1&paramN=valueN]
 */
func (optional *DBOptional) GenerateConfig() string {
	if optional.DBCharset == "" {
		optional.DBCharset = "utf8"
	}

	config := ""
	if !isConsulName(optional.DBHostname) {
		format := "%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local&timeout=%s&readTimeout=%s&writeTimeout=%s"
		config = fmt.Sprintf(format, optional.User, optional.Password, optional.DBHostname, optional.DBPort,
			optional.DBName, optional.DBCharset, optional.Timeout, optional.ReadTimeout, optional.WriteTimeout)
	} else {
		format := "%s:%s@tcp(%s)/%s?charset=%s&parseTime=True&loc=Local&timeout=%s&readTimeout=%s&writeTimeout=%s"
		config = fmt.Sprintf(format, optional.User, optional.Password, optional.DBHostname,
			optional.DBName, optional.DBCharset, optional.Timeout, optional.ReadTimeout, optional.WriteTimeout)
	}

	//optional.DBName, optional.Timeout)
	if optional.ExtraDSNMap != nil && len(optional.ExtraDSNMap) != 0 {
		for confKey, confVal := range optional.ExtraDSNMap {
			config += fmt.Sprintf("&%s=%s", confKey, confVal)
		}
	}

	return config
}

func isConsulName(name string) bool {
	// is a P.S.M
	tmp := strings.Split(name, ".")
	return len(tmp) == 3 || len(tmp) == 5
}

var db = make(map[string]DBOptional)
var dbLock sync.RWMutex

func InitDBConf(filename string) error {
	dbBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	dbLock.Lock()
	defer dbLock.Unlock()
	if err = yaml.Unmarshal(dbBytes, &db); err != nil {
		logs.Error("InitDBConf %v", err)
		return err
	}
	// Override with ssconf
	for k, v := range db {
		if strings.HasSuffix(v.DriverName, ".conf") && v.DBName != "" {
			if optional := getFromSsConf(v.DriverName, v.DBName, v.DBHostname); optional.DBHostname != "" {
				db[k] = optional
			}
		}
	}
	return nil
}

func getFromSsConf(filename string, db string, cluster string) DBOptional {
	ret, err := ssconf.LoadSsConfFile(filename)
	if err != nil {
		logs.Error("LoadSsConfFile error %v", err)
		return DBOptional{}
	}
	return GetDbConf(ret, db, cluster)
}

func DBConf(dbName string) (*DBOptional, error) {
	dbLock.RLock()
	defer dbLock.RUnlock()
	if c, ok := db[dbName]; ok {
		return &c, nil
	} else {
		return nil, fmt.Errorf("Can't find DB with name %s", dbName)
	}
}

type ThroughCacheConfig struct {
	// mc cluster name
	Cluster string
	// mc服务列表['ip:port', ...]
	ServerList []string `yaml:"ServerList"`
	// mc读写超时时间
	Timeout time.Duration `yaml:"Timeout"`
	// 缓存失效时间，单位s
	ExpireTime int32 `yaml:"ExpireTime"`
	// 缓存key的格式
	KeyFormat string `yaml:"KeyFormat"`
	// 缓存版本
	Version string `yaml:"Version"`
	// 设置阻塞或者异步方式更新缓存
	BlockThrough bool `yaml:"BlockThrough"`
}

var cache = make(map[string]*ThroughCacheConfig)
var cacheLock sync.RWMutex

func InitCacheConf(filename string) error {
	cacheBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	cacheLock.Lock()
	defer cacheLock.Unlock()
	if err = yaml.Unmarshal(cacheBytes, &cache); err != nil {
		logs.Error("InitCacheConf %v", err)
		return err
	}
	for name, c := range cache {
		logs.Info("Name: %s, Cache: %v", name, c)
	}
	return nil
}

func CacheConf(cacheName string) (*ThroughCacheConfig, error) {
	cacheLock.RLock()
	defer cacheLock.RUnlock()

	if c, ok := cache[cacheName]; ok {
		return c, nil
	} else {
		return nil, fmt.Errorf("Can't find cache with name %s", cacheName)
	}
}
