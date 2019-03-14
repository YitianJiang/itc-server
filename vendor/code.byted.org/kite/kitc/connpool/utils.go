package connpool

import (
	"code.byted.org/gopkg/metrics"
	"code.byted.org/gopkg/env"
)

var (
	mclient *metrics.MetricsClient
)

func init() {
	psm := env.PSM()
	if psm == env.PSMUnknown {
		psm = "toutiao.unknown.unknown"
	}
	prefix := "go." + psm

	mclient = metrics.NewDefaultMetricsClient(prefix, true)
}

// 受限于基础组件, 当下游addr太多时, opentsdb聚合会很慢甚至聚合不出来, 故暂时不把addr放到tag内
func getShortConnSucc(targetPSM, addr string) {
	mclient.EmitCounter("to."+targetPSM+".get_short_conn_succ", 1, "", nil)
}

func getShortConnFail(targetPSM, addr string) {
	mclient.EmitCounter("to."+targetPSM+".get_short_conn_fail", 1, "", nil)
}

func newLongConnSucc(targetPSM, addr string) {
	mclient.EmitCounter("to."+targetPSM+".new_long_conn_succ", 1, "", nil)
}

func newLongConnFail(targetPSM, addr string) {
	mclient.EmitCounter("to."+targetPSM+".new_long_conn_fail", 1, "", nil)
}

func getAliveLongConnSucc(targetPSM, addr string) {
	mclient.EmitCounter("to."+targetPSM+".get_alive_long_conn_succ", 1, "", nil)
}

func longConnError(targetPSM, addr string) {
	mclient.EmitCounter("to."+targetPSM+".long_conn_error", 1, "", nil)
}

func putLongConnSucc(targetPSM, addr string) {
	mclient.EmitCounter("to."+targetPSM+".put_long_conn_succ", 1, "", nil)
}

func putLongConnFail(targetPSM, addr string) {
	mclient.EmitCounter("to."+targetPSM+".put_long_conn_fail", 1, "", nil)
}
