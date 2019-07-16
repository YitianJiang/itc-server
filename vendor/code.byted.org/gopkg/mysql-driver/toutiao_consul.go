package mysql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"code.byted.org/gopkg/asyncache"
	"code.byted.org/inf/infsecc"
)

/*
	There are four patterns of DSN in toutiao now:
		1) consul with prefix "consul:":
			USERNAME:PASSWORD@tcp(consul:toutiao.mysql.ms_data_write)/DATABASE
		2) consul without prefix:
			USERNAME:PASSWORD@tcp(toutiao.mysql.ms_data_write)/DATABASE
		3) multi-host DSN:
			USERNAME:PASSWORD@tcp(10.4.16.18:3306,127.0.0.1:3306)/DATABASE
		4) normal DSN(single host):
			USERNAME:PASSWORD@tcp(10.4.16.18:3306)/DATABASE
		5) multi-host-one-port:
			USERNAME:PASSWORD@tcp(10.4.16.18,127.0.0.1:3306)/DATABASE
		6) unix socket file:

''
	convertConsulDSN Convert pattern 1 and 2 to pattern 3 or 4, and return consulName;
*/
const (
	consulPrefix         = "consul:"
	dbauthService        = "toutiao.mysql.dbauth_service"
	dbauthServiceTestEnv = "toutiao.mysql.dbauth_service_testenv"
)

var (
	MeshSwithch    string
	MeshSocketPath string
	AuthSwitch     string
	Mode           string = "Normal"
)

func consulName2EnvKey(s string) string {
	s = strings.Replace(s, ".", "_", -1)
	s = strings.ToUpper(s) + "_AUTHKEY"
	return s
}

func convertConsulDSN(dsn string) (converedDSN string, consulName string) {
	var requser string
	var varifyToken = false
	var varifyAuthkey = false
	originDSN := dsn

	hookTag := "@tcp("
	left := strings.Index(dsn, hookTag)
	if left == -1 {
		return originDSN, ""
	}

	authStr := dsn[:left] // for parse psm/authkey

	left += len(hookTag)

	if strings.HasPrefix(dsn[left:], consulPrefix) {
		// pattern 1, remove prefix
		dsn = dsn[:left] + dsn[left+len(consulPrefix):]
	}

	right := strings.Index(dsn[left:], ")")

	//dbinformation
	dbleft := strings.Index(dsn, "/")
	dbright := strings.Index(dsn, "?")
	if right == -1 {
		return originDSN, ""
	}
	right += left

	if isInvalidPSM(dsn[left:right]) == false {
		str := dsn[left:right]
		if len(strings.Split(str, ",")) > 1 && len(strings.Split(str, ":")) == 2 {
			// pattern 5, convert it to pattern 3
			tmp := strings.Split(str, ":")
			port := tmp[1]
			hosts := strings.Split(tmp[0], ",")
			addrs := make([]string, 0, len(hosts))
			for _, host := range hosts {
				addrs = append(addrs, fmt.Sprintf("%v:%v", host, port))
			}
			addrStr := strings.Join(addrs, ",")
			return dsn[:left] + addrStr + dsn[right:], ""
		}
		// pattern 3 or 4, return directly
		return originDSN, ""
	}

	consulName = dsn[left:right]
	//convert to mesh module with socket
	// tcp -> unix  env control && socket file  && user opt on dsn
	if MeshSwithch == "True" && MeshSocketPath != "" {
		if !strings.Contains(dsn, "disableMesh=true") {
			return consulName + ":" + "" + "@unix(" + MeshSocketPath + dsn[right:dbleft] + dsn[dbleft:dbright] + dsn[dbright:], consulName
		}
	}
	var addrs []ConsulEndpoint
	var err error
	for try := 0; try < 3; try++ {
		addrs, err = translateOne(consulName)
		if err == nil {
			break
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "[mysql-driver]: consul translate %v err: %v \n", consulName, err)
		return originDSN, ""
	}

	addrList := make([]string, 0, len(addrs))
	for _, end := range addrs {
		if end.Host != "" {
			addrList = append(addrList, fmt.Sprintf("%v:%v", end.Host, end.Port))
		}
	}
	if len(addrList) == 0 {
		fmt.Fprintf(os.Stderr, "[mysql-driver]: no host found for consulName: %v \n", consulName)
		return originDSN, ""
	}
	addrsStr := strings.Join(addrList, ",")

	token, err := infsecc.GetToken(true)
	if err != nil {
		// fmt.Fprintf(os.Stderr, "[mysql-driver]: GetToken error :%s \n", err.Error())
	}
	authKey := os.Getenv(consulName2EnvKey(consulName))
	// parse psm/authkey
	parts := strings.Split(authStr, ":")
	if len(parts) == 2 {
		requser = parts[0]
		if parts[0] != "" && isInvalidPSM(parts[0]) {
			serviceName = parts[0]
		}
		if parts[1] != "" && isInvalidPSM(parts[0]) {
			authKey = parts[1]
		}
	}
	if authKey != "" && isInvalidPSM(serviceName) && serviceName != "toutiao.unknown.unknown" {
		varifyAuthkey = true
	}
	if (isInvalidPSM(requser) || requser == "") && isInvalidPSM(consulName) && token != "" {
		varifyToken = true
	}
	if AuthSwitch == "1" && token != "" {
		varifyToken = true
	}
	// if (authKey != "" || isInvalidPSM(requser)) && isInvalidPSM(serviceName) && serviceName != "toutiao.unknown.unknown"
	if varifyAuthkey || varifyToken {
		dbinfo, err := getDbInfoFormAuthModule(serviceName, consulName, authKey)
		if dbinfo != nil && err == nil {
			if dbright != -1 { // has parameter
				return dbinfo.user + ":" + dbinfo.pwd + hookTag + addrsStr + dsn[right:dbleft+1] + dbinfo.dbname + dsn[dbright:], consulName
			} else { // doesn't has any paramenter
				return dbinfo.user + ":" + dbinfo.pwd + hookTag + addrsStr + dsn[right:dbleft+1] + dbinfo.dbname, consulName
			}
		}
		fmt.Fprintf(os.Stderr, "[mysql-driver]:  failed get auth info %v ,err: %v \n", dbinfo, err)
	}
	return dsn[:left] + addrsStr + dsn[right:], consulName
}

func isInvalidPSM(psm string) bool {
	segNum := len(strings.Split(psm, "."))
	return segNum == 3 || segNum == 5
}

func addrToConsulName(addr string) string {
	tmp := strings.Split(addr, ":")
	if len(tmp) != 2 {
		return addr
	}
	return strings.Replace(tmp[1], ".", "_", -1)
}

var consulCache *asyncache.Asyncache
var authCache *asyncache.Asyncache

// type Getter func(key string) (interface{}, error)
// Options .

func init() {
	AuthSwitch = os.Getenv("SEC_MYSQL_AUTH")
	MeshSwithch = os.Getenv("TCE_ENABLE_MYSQL_SIDECAR_EGRESS")
	MeshSocketPath = os.Getenv("SERVICE_MESH_MYSQL_ADDR")
	if MeshSwithch == "True" && MeshSocketPath != "" {
		Mode = "MysqlMesh"
		fmt.Fprintf(os.Stdout, "[mysql-driver] Start with Mesh mode  sockefile %s\n", MeshSocketPath)
	} else {
		Mode = "Normal"
		fmt.Fprintf(os.Stdout, "[mysql-driver] Start with Normal mode consul/psmauth/user-passwd\n")
	}
	if AuthSwitch == "1" {
		fmt.Fprintf(os.Stdout, "[mysql-driver] init with authSwitch on \n")
	}
	consulgetter := func(key string) (interface{}, error) {
		eps, err := consulGet(key)
		if err != nil {
			return nil, err
		}
		return eps, nil
	}
	consulErr := func(key string, err error) {
		if err != nil {
			fmt.Fprintf(os.Stderr, "[mysql-driver]: consulCache [%s] error %s \n", key, err.Error())
		}
	}
	consulOpt := asyncache.Options{BlockIfFirst: true, RefreshDuration: time.Second * 10, Fetcher: consulgetter, ErrHandler: consulErr}

	consulCache = asyncache.NewAsyncache(consulOpt)

	authgetter := func(key string) (interface{}, error) {
		lists := strings.Split(key, "#")
		if len(lists) != 3 {
			return nil, fmt.Errorf("auth req format err %s", key)
		}
		user_pwd, err := getServiceInfo(lists[0], lists[1], lists[2])
		if err != nil {
			return nil, err
		}
		return user_pwd, nil
	}
	authlErr := func(key string, err error) {
		if err != nil {
			fmt.Fprintf(os.Stderr, "[mysql-driver]: authCache [%s] error %s", key, err.Error())
		}
	}
	authOpt := asyncache.Options{BlockIfFirst: true, RefreshDuration: time.Second * 120, Fetcher: authgetter, ErrHandler: authlErr}
	authCache = asyncache.NewAsyncache(authOpt)
}

func translateOne(consulName string) ([]ConsulEndpoint, error) {
	var val interface{}
	var err error
	val = consulCache.Get(consulName, nil)
	if val == nil && consulName != dbauthService { // dbauthService use dsn instead
		val, err = getServiceHost(consulName)
		if err != nil {
			return nil, err
		}
	}
	eps, ok := val.([]ConsulEndpoint)
	if !ok {
		// return eps, nil
		return nil, fmt.Errorf("translateOne consulName err: invalid val type")
	}
	if len(eps) == 0 && consulName != dbauthService {
		val, err = getServiceHost(consulName)
		if err != nil {
			return nil, err
		}
		eps, ok = val.([]ConsulEndpoint)
		if !ok {
			return nil, fmt.Errorf("translateOne consulName err: invalid val type")
		}
	}
	return eps, nil
}

type Dbinfo struct {
	user   string
	pwd    string
	dbname string
}

func getDbInfoFormAuthModule(serviceName, consulName, authKey string) (*Dbinfo, error) {
	key := fmt.Sprintf("%s#%s#%s", serviceName, consulName, authKey)
	item := authCache.Get(key, nil)
	if item == nil {
		return nil, fmt.Errorf("get info from cache error")
	}
	switch v := item.(type) {
	case string:
		tmp := strings.Split(v, "-")
		if len(tmp) != 3 {
			return nil, fmt.Errorf("cache info error: %v", v)
		}
		return &Dbinfo{
			user:   tmp[0],
			pwd:    tmp[1],
			dbname: tmp[2],
		}, nil
	default:
		return nil, fmt.Errorf("Cache format error: %v", v)
	}
}

type DbInfoReq struct {
	ServiceName string `json:"serviceName"`
	// Key represents the unique location of this Node (e.g. "/foo/bar").
	Psm     string `json:"psm"`
	Token   string `json:"token"`
	AuthKey string `json:"authkey"`
}

// getServiceInfo get username,passwork,dbaname from backend dbauth service ,
//check if online or offline service use subfix testenv
func getServiceInfo(serviceName, consulName, authKey string) (string, error) {
	var err error
	var start = time.Now().UnixNano() / 1e3
	var metricsInfo Metrics_Info
	metricsInfo.Psm = serviceName
	metricsInfo.ServiceName = consulName
	token, err := infsecc.GetToken(true)
	if err != nil {
		// fmt.Fprintf(os.Stderr, "[mysql-driver]: GetToken error :%s \n", err.Error())
	}
	defer func() {
		metricsInfo.Cost = time.Now().UnixNano()/1e3 - start
		doAuthMetrics(&metricsInfo)
	}()
	var host string
	var port int
	var url string
	var dbauthservice string
	if strings.HasSuffix(consulName, "testenv") { // test env db consul name
		dbauthservice = dbauthServiceTestEnv
	} else {
		dbauthservice = dbauthService
	}
	hosts, err := translateOne(dbauthservice)
	if len(hosts) > 0 && err == nil {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		index := r.Intn(len(hosts))
		host = hosts[index].Host
		port = hosts[index].Port
		url = fmt.Sprintf("http://%s:%d/getdbinfo", host, port)
	} else {
		fmt.Fprintf(os.Stderr, "[mysql-driver]: no host found for consulName: %v ,err: %v\n", "toutiao.mysql.dbauth_service", err)
		if dbauthservice == dbauthServiceTestEnv {
			return "", fmt.Errorf("consul not availble for testenv")
		}
		url = fmt.Sprintf("http://dbauth.byted.org/getdbinfo")
	}
	metricsInfo.Host = host
	cont, err := json.Marshal(&DbInfoReq{ServiceName: consulName, Psm: serviceName, AuthKey: authKey, Token: token})
	if err != nil {
		metricsInfo.ErrCode = 1
		return "", err
	}
	u_p_d, err := post(url, cont) //Fetch_data(host, port, consulName, serviceName, authKey)
	if err != nil {
		metricsInfo.ErrCode = 2
	}
	if err = checkAuthinfo(u_p_d); err != nil {
		return "", err
	}
	return u_p_d, nil

}

//func check return authinfo
func checkAuthinfo(u_p_d string) error {
	if u_p_d == "" {
		return fmt.Errorf("Auth failed")
	}
	ll := strings.Split(u_p_d, "-")
	if len(ll) != 3 {
		return fmt.Errorf("Auth failed")
	}
	if len(ll[0]) == 0 || len(ll[1]) == 0 {
		return fmt.Errorf("Auth failed")
	}
	return nil
}

func getServiceHost(consulName string) ([]ConsulEndpoint, error) {
	var err error
	var start = time.Now().UnixNano() / 1e3
	var metricsInfo Metrics_Info

	metricsInfo.Psm = serviceName
	metricsInfo.ServiceName = consulName
	defer func() {
		metricsInfo.Cost = time.Now().UnixNano()/1e3 - start
		doAuthMetrics(&metricsInfo)
	}()
	url := fmt.Sprintf("http://dbauth.byted.org/getserviceip?servicename=%s", consulName)

	resp, err := get(url) //Fetch_data(host, port, consulName, serviceName, authKey)
	if err != nil {
		metricsInfo.ErrCode = 1
		return nil, err
	}
	if resp == "" {
		metricsInfo.ErrCode = 2
		return nil, fmt.Errorf("get service ip err")
	}
	eps := make([]ConsulEndpoint, 0)
	hosts := strings.Split(resp, "-") //ip:port,ip:port....
	if len(hosts) <= 0 {
		return nil, fmt.Errorf("get serviceerr")
	}

	for _, host := range hosts {
		tmp := strings.Split(host, ":")
		port, err := strconv.Atoi(tmp[1])
		if err != nil {
			continue
		}
		ep := ConsulEndpoint{Host: tmp[0], Port: port}
		eps = append(eps, ep)
	}
	return eps, nil
}

var (
	transport = &http.Transport{DisableKeepAlives: true}
	client    = &http.Client{Transport: transport, Timeout: 500 * time.Millisecond}
)

func get(url string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func post(url string, cont []byte) (string, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(cont))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
