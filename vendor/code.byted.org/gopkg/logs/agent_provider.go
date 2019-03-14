package logs

import (
	"os"
	"strconv"

	"code.byted.org/log_market/gosdk"
)

const (
	canaryTaskName = "_canary"
	rpcTaskName    = "_rpc"
)

// AgentProvider : logagent provider, implement KVLogProvider
type AgentProvider struct {
	level    int
	isRPCLog bool
	pid      string
}

// NewAgentProvider : factory for AgentProvider.
func NewAgentProvider() *AgentProvider {
	return &AgentProvider{isRPCLog: false}
}

// NewRPCLogAgentProvider :
// 	if this provider is used to deal with RPC log, set isRPCLog true.
// 	only used for RPC log (in in kite/log.go).
func NewRPCLogAgentProvider() *AgentProvider {
	return &AgentProvider{isRPCLog: true}
}

// Init : implement KVLogProvider
func (ap *AgentProvider) Init() error {
	ap.pid = strconv.Itoa(os.Getpid())
	return nil
}

// SetLevel : implement KVLogProvider
func (ap *AgentProvider) SetLevel(level int) {
	ap.level = level
}

// WriteMsg : log agent do not support write message any more
func (ap *AgentProvider) WriteMsg(msg string, level int) error {
	// use WriteMsgKVs instead
	return nil
}

// WriteMsgKVs : implement KVLogProvider, core method for this provider
func (ap *AgentProvider) WriteMsgKVs(level int, msg string, headers map[string]string, kvs map[string]string) error {
	if level < ap.level {
		return nil
	}

	// same with python - https://review.byted.org/#/c/toutiao/frame/+/613211/5/streamlog/formatter.py

	kvs["_level"] = headers["level"]
	kvs["_ts"] = headers["timestamp"]
	kvs["_host"] = headers["host"]
	kvs["_language"] = "go"
	kvs["_taskName"] = headers["psm"]
	kvs["_psm"] = headers["psm"]
	kvs["_cluster"] = headers["cluster"] // 对于RPC日志的最终KVs来说，cluster 是当前服务的集群，_cluster是远程服务的集群……不过这样设置真的OK么？
	kvs["_logid"] = headers["logid"]
	kvs["_deployStage"] = headers["stage"]
	kvs["_podName"] = headers["pod_name"]
	kvs["_process"] = ap.pid
	kvs["_version"] = string(versionBytes) // "v1(6)"
	kvs["_location"] = headers["location"]

	message := &gosdk.Msg{
		Msg:  []byte(msg),
		Tags: kvs,
	}

	if ap.isRPCLog {
		gosdk.Send(rpcTaskName, message)
		return nil
	}

	gosdk.Send(kvs["_taskName"], message)

	return nil
}

// Destroy : implement KVLogProvider
func (ap *AgentProvider) Destroy() error {
	return nil
}

// Flush : implement KVLogProvider
func (ap *AgentProvider) Flush() error {
	return nil
}
