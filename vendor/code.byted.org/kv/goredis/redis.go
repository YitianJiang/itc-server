package goredis

import (
	"context"
	"time"

	"code.byted.org/gopkg/logs"
	redis "code.byted.org/kv/redis-v6"
	"code.byted.org/kv/redis-v6/pkg/pool"
)

type Client struct {
	*redis.Client

	cluster            string
	psm                string
	metricsServiceName string
	ctx                context.Context
	clusterpool        *MultiServPool /* cluster connpool */
}

// NewClient will create a new client with cluster name use the default timeout settings
func NewClient(cluster string) (*Client, error) {
	opt := NewOption()
	return NewClientWithOption(cluster, opt)
}

// NewClientWithOption will use user specified timeout settings in option
func NewClientWithOption(cluster string, opt *Option) (*Client, error) {
	servers, err := loadConfByClusterName(cluster, opt.configFilePath, opt.useConsul)
	if err != nil {
		return nil, err
	}
	logs.Info("Cluster %v's server list is %v", cluster, servers)
	return NewClientWithServers(cluster, servers, opt)
}

// NewClientWithServers will create a new client with specified servers and timeout in option
func NewClientWithServers(cluster string, servers []string, opt *Option) (*Client, error) {
	if len(servers) == 0 {
		return nil, ErrEmptyServerList
	}
	serversCh := make(chan []string, 1)
	serversCh <- servers

	cli := &Client{
		Client:             redis.NewClient(opt.Options),
		cluster:            GetClusterName(cluster),
		psm:                checkPsm(),
		metricsServiceName: GetPSMClusterName(cluster),
		clusterpool:        NewMultiServPool(servers, serversCh, opt),
	}
	cli.WrapProcess(cli.metricsWrapProcess)
	cli.WrapGetConn(cli.GetConn)
	cli.WrapReleaseConn(cli.ReleaseConn)

	//pre conn
	preidx := make([]int, opt.PoolInitSize)
	preconn := make([]*pool.Conn, opt.PoolInitSize)

	for i := range preidx {
		preconn[i], _, _ = cli.GetConn()
	}
	for _, cn := range preconn {
		if cn != nil {
			cli.ReleaseConn(cn, nil)
		}
	}
	isInWhiteList(cluster)

	if opt.autoLoadConf {
		autoLoadConf(cli.cluster, serversCh, opt)
	}

	return cli, nil
}

func (c *Client) clone() *Client {
	return &Client{
		Client:             c.Client,
		cluster:            c.cluster,
		psm:                c.psm,
		metricsServiceName: c.metricsServiceName,
	}
}

// WithContext .
func (c *Client) WithContext(ctx context.Context) *Client {
	cc := c.clone()
	cc.WrapProcess(cc.metricsWrapProcess)
	cc.ctx = ctx
	return cc
}

/* get conn from multi servs pool */
func (c *Client) GetConn() (*pool.Conn, bool, error) {
	cn, isNew, err := c.clusterpool.getConnection()
	if err != nil {
		return nil, false, err
	}
	//need init
	if !cn.Inited {
		if err := c.Client.InitConn(cn); err != nil {
			_ = c.ReleaseConn(cn, err)
			return nil, false, err
		}
	}
	return cn, isNew, nil
}

/* release conn to multi servs connpool, bad conn->remove done conn->put to connpool */
func (c *Client) ReleaseConn(cn *pool.Conn, err error) bool {
	return c.clusterpool.releaseConnection(cn, err)
}

func (c *Client) metricsWrapProcess(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
	return func(cmd redis.Cmder) error {
		// degradate
		/*
			if cmdDegredated(c.metricsServiceName, cmd.Name()) {
				cmd.SetErr(ErrDegradated)
				return ErrDegradated
			}
		*/
		// if stress rpc, hack args
		if prefix, ok := isStressTest(c.ctx); ok {
			cmd = convertStressCMD(prefix, cmd)
		}
		t0 := time.Now().UnixNano()
		err := oldProcess(cmd)
		latency := (time.Now().UnixNano() - t0) / 1000
		addCallMetrics(c.ctx, cmd.Name(), latency, err, c.cluster, c.psm, c.metricsServiceName, 1)

		return err
	}
}
func (c *Client) Pipeline() *Pipeline {
	pipe := c.NewPipeline("pipeline")
	return pipe
}

// this func will create a pipeline with name user specified
// the name will be used for pipeline metrics
func (c *Client) NewPipeline(pipelineName string) *Pipeline {
	pipe := &Pipeline{
		c.Client.Pipeline(),
		c,
		c.cluster,
		c.psm,
		c.metricsServiceName,
		pipelineName,
	}
	return pipe
}
