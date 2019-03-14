package etcdutil

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

var (
	agentHost = "127.0.0.1"
	agentPort = "2150"
	agentAddr = ""
	useAgent  = false

	pingTimeout  = 50 * time.Millisecond
	pingSuccCode = 0
)

type response struct {
	ErrCode int    `json:"errorCode"`
	Message string `json:"message"`
}

func init() {
	if host := os.Getenv("TCE_HOST_IP"); host != "" {
		agentHost = host
	}
}

func getAgentAddr() string {
	return agentHost + ":" + agentPort
}

func pingAgent() bool {
	client := &http.Client{Timeout: pingTimeout}
	pingURL := "http://" + agentHost + ":" + agentPort + "/agent/ping"
	agentResp, err := client.Get(pingURL)
	if err != nil {
		return false
	}
	defer agentResp.Body.Close()

	var resp response
	if err := json.NewDecoder(agentResp.Body).Decode(&resp); err != nil {
		return false
	}
	if resp.ErrCode == pingSuccCode {
		return true
	}
	return false
}

func init() {
	go func() {
		interval := time.Second * 1
		MAX := time.Second * 60
		for {
			if pingAgent() {
				useAgent = true
				agentAddr = getAgentAddr()
				cli, err := NewClient()
				if err != nil {
					fmt.Println("ETCD: reset default client err: ", err.Error())
					return
				}

				fmt.Println("ETCD: use bagent to access ETCD")
				defaultClientMu.Lock()
				defaultClient = cli
				defaultClientMu.Unlock()
				return
			}

			metricsClient.EmitStore("agent.cannot.connect", 1, "etcd", langTag)
			time.Sleep(interval)
			interval = interval * 2
			if interval > MAX {
				interval = MAX
			}
		}
	}()
}
