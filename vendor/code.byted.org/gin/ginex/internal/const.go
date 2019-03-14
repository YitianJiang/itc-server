package internal

const (
	TT_LOGID_HEADER_KEY          = "X-TT-LOGID"    // Http header中log id key
	TT_LOGID_HEADER_FALLBACK_KEY = "X-Tt-Logid"    // unknown fallback
	TT_ENV_KEY                   = "X-TT-ENV"      // env in http header
	TT_TRACE_TAG                 = "X-TT-TRACE"    // force trace tag in http header
	TT_STRESS_KEY                = "X-Tt-Stress"   // 压测流量标识
	LOGIDKEY                     = "K_LOGID"       // 唯一的Request ID
	SNAMEKEY                     = "K_SNAME"       // 本服务的名字
	LOCALIPKEY                   = "K_LOCALIP"     // 本服务的IP 地址
	CLUSTERKEY                   = "K_CLUSTER"     // 本服务集群的名字
	METHODKEY                    = "K_METHOD"      // 本服务当前所处的接口名字（也就是Method名字）
	ENVKEY                       = "K_ENV"         // 传递给Kite的Env
	STRESSKEY                    = "K_STRESS"      // 传递给Kite的压测标识
	BDIFFCTXKEY                  = "bdiff_ctx_key" // 传递给bdiff的参数
	HOST_IP_ADDR                 = "HOST_IP_ADDR"
	TCE_CLUSTER                  = "TCE_CLUSTER"
	SERVICE_CLUSTER              = "SERVICE_CLUSTER"
)

// Envs set by ginex, begins with "_GINEX"
const (
	GINEX_PSM = "_GINEX_PSM"
)

// Ginex framework version
const (
	VERSION = "v1.1.2"
)
