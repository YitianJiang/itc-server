package kite

/*
   提供端口供pprof
*/

import (
	"encoding/json"
	"net/http"
	_ "net/http/pprof"

	"code.byted.org/gopkg/logs"
	"code.byted.org/kite/kitc"
)

// RegisterDebugHandler add custom http interface and handler
func RegisterDebugHandler(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	http.HandleFunc(pattern, handler)
}

func startDebugServer() {
	if !EnableDebugServer {
		logs.Info("KITE: Debug server not enabled.")
		return
	}

	RegisterDebugHandler("/version", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(ServiceVersion))
	})

	RegisterDebugHandler("/runtime/kite", runtimeKite)
	RegisterDebugHandler("/runtime/kitc", runtimeKitc)
	RegisterDebugHandler("/idl/kite", idlKite)
	RegisterDebugHandler("/idl/kitc", idlKitc)

	go func() {
		logs.Info("KITE: Start pprof listen on: %s", DebugServerPort)
		// Use default mux make easy for new url path like trace.
		err := http.ListenAndServe(DebugServerPort, nil)
		if err != nil {
			logs.Noticef("KITE: Start debug server failed: %s", err)
		}
	}()
}

func runtimeKite(w http.ResponseWriter, r *http.Request) {
	confs := RPCServer.RemoteConfigs()
	connNow, connLim, qpsLim := RPCServer.Overload()
	meta := RPCServer.Metainfo()

	m := map[string]interface{}{
		"remote_configs": confs,
		"metainfo":       meta,
		"overload": map[string]interface{}{
			"conn_now": connNow,
			"conn_lim": connLim,
			"qps_lim":  qpsLim,
		},
	}
	buf, _ := json.MarshalIndent(m, "", "  ")
	w.WriteHeader(200)
	w.Write(buf)
}

func runtimeKitc(w http.ResponseWriter, r *http.Request) {
	clients := kitc.AllKitcClients()
	results := make(map[string]interface{})
	for location, client := range clients {
		m := map[string]interface{}{
			"name":                 client.Name(),
			"downstream_instances": client.ServiceInstances(),
			"remote_configs":       client.RemoteConfigs(),
			"recent_events":        client.RecentEvents(),
			"options":              client.Options(),
		}

		serviceCBMap := make(map[string]interface{})
		serviceCBs := client.ServiceCircuitbreaker()
		for service, breaker := range serviceCBs.DumpBreakers() {
			serviceCBMap[service] = map[string]interface{}{
				"state":             breaker.State().String(),
				"successes in 10s":  breaker.Successes(),
				"failures in 10s":   breaker.Failures(),
				"timeouts in 10s":   breaker.Timeouts(),
				"error rate in 10s": breaker.ErrorRate(),
			}
		}
		m["service_circuitbreaker"] = serviceCBMap

		instanceCBMap := make(map[string]interface{})
		instanceCBs := client.InstanceCircuitbreaker()
		for ip, breaker := range instanceCBs.DumpBreakers() {
			instanceCBMap[ip] = map[string]interface{}{
				"state":             breaker.State().String(),
				"successes in 10s":  breaker.Successes(),
				"failures in 10s":   breaker.Failures(),
				"timeouts in 10s":   breaker.Timeouts(),
				"error rate in 10s": breaker.ErrorRate(),
			}
		}
		m["ip_circuitbreaker"] = instanceCBMap

		results[location] = m
	}
	buf, _ := json.MarshalIndent(results, "", "  ")
	w.WriteHeader(200)
	w.Write(buf)
}

func idlKite(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	for fname, content := range IDLs {
		w.Write([]byte("==== " + fname + ":\n"))
		w.Write([]byte(content))
	}
}

func idlKitc(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	for service, m := range kitc.IDLs {
		for fname, content := range m {
			w.Write([]byte("==== " + service + " " + fname + ":\n"))
			w.Write([]byte(content))
		}
	}
}
