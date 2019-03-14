package ginex

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"code.byted.org/gopkg/logs"
	"code.byted.org/hystrix/hystrix-go/hystrix"
)

var (
	debugMux = http.NewServeMux()
)

// 业务代码注册debug handler
func RegisterDebugHandler(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	debugMux.HandleFunc(pattern, handler)
}

func startDebugServer() {
	if !appConfig.EnablePprof {
		logs.Info("Debug server not enabled.")
		return
	}
	if appConfig.DebugPort == 0 {
		logs.Info("Debug port is not specified.")
		return
	}

	// pprof handler
	debugMux.HandleFunc("/debug/pprof/", pprof.Index)
	debugMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	debugMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	debugMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	debugMux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// hystrix handler
	hystrixStreamHandler := hystrix.NewStreamHandler()
	hystrixStreamHandler.Start()
	debugMux.Handle("/debug/hystrix.stream", hystrixStreamHandler)

	go func() {
		debugPort := appConfig.DebugPort
		logs.Infof("Start pprof and hystrix listen on: %d", debugPort)
		err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", debugPort), debugMux)
		if err != nil {
			logs.Fatalf("Failed to start debug server: %s", err)
		}
	}()
}
