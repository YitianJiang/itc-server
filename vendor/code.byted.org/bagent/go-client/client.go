package bagentutil

import (
	"fmt"
	"net"
	"time"

	"code.byted.org/bagent/go-client/thrift_gen/agent"
	"code.byted.org/bagent/go-client/thrift_gen/base"
	"code.byted.org/gopkg/thrift"
)

const (
	agentAddr    = "/tmp/prod_bagent.sock"
	devAgentAddr = "/tmp/dev_bagent.sock"
)

// Client for agentutil
type Client struct {
	transport   thrift.TTransport
	agentClient *agent.AgentServiceClient
}

func testClient() (*Client, error) {
	return newClient(devAgentAddr)
}

// NewClient create a new Client
func NewClient() (*Client, error) {
	return newClient(agentAddr)
}

func newClient(addr string) (*Client, error) {
	conn, err := net.DialTimeout("unix", addr, time.Duration(300*time.Millisecond))
	if err != nil {
		return nil, fmt.Errorf("Connect to agent error: %s", err)
	}
	socket := thrift.NewTSocketFromConnTimeout(conn, time.Duration(300*time.Millisecond))
	transportFactory := thrift.NewTBufferedTransportFactory(4096)
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	transport := transportFactory.GetTransport(socket)
	return &Client{
		transport:   transport,
		agentClient: agent.NewAgentServiceClientFactory(transport, protocolFactory),
	}, nil
}

// Close the client if any error
// If agent restart, we must re-create the client.
func (client *Client) Close() {
	client.transport.Close()
}

// ReportInfo report server's metadata
func (client *Client) ReportInfo(infos map[string]string) error {
	req := agent.NewFetchUploadInfosRequest()
	req.Infos = infos

	base := base.NewBase()
	base.Caller = infos["psm"]
	req.Base = base

	resp, err := client.agentClient.ReportInfo(req)
	if err != nil || resp == nil {
		return fmt.Errorf("Report info error: %s", err)
	}

	return nil
}

func (client *Client) GetTonfig(callerPsm, path, version string) (map[string]string, error) {
	req := agent.NewGetTonfigRequest()
	req.KeyPath = path
	req.Version = version
	base := base.NewBase()
	base.Caller = callerPsm
	req.Base = base

	resp, err := client.agentClient.GetTonfig(req)
	if resp != nil {
		return resp.GetKvs(), err
	}
	return make(map[string]string), err
}
