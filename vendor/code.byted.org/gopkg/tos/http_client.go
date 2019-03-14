package tos

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"
)

const clusterSep = "$"
const maxConnRetryTimes = 2

var dialer = net.Dialer{Timeout: 3 * time.Second}

// 1. use only conn succ ratio to decide quality
//    when get one addr, connect start, connect done.
// 2. how to do retry, how to count retry times?
//    use ctx

type httpClient struct {
	client  http.Client
	addrMan *addrManager
	cluster string
}

func isIPAddr(s string) bool {
	h, _, _ := net.SplitHostPort(s)
	return h != "" && net.ParseIP(h) != nil
}

// NewHttpClient returns HttpClient with `http://{YOUR_SERVICE}/path/to/your/api` support
func newHttpClient(cluster, idc string) (*httpClient, error) {
	addrMan, err := newAddrManager(cluster, idc)
	if err != nil {
		return nil, err
	}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if idx := strings.Index(addr, clusterSep); idx > 0 {
				addr = addr[:idx] // rm cluster name
			}
			if isIPAddr(addr) {
				return dialer.DialContext(ctx, network, addr)
			}
			var err error
			var conn net.Conn
			for i := 0; i < maxConnRetryTimes; i++ {
				addr := addrMan.getAddr()
				conn, err = dialer.DialContext(ctx, network, addr)
				if err == nil {
					addrMan.cntSucc(addr)
					return conn, nil
				} else {
					if strings.Contains(err.Error(), "connection refused") {
						addrMan.fastCntFail(addr)
					} else {
						addrMan.cntFail(addr)
					}
				}
			}

			return nil, fmt.Errorf("retry %d times, still cannot connect to server, last conn err: %s", maxConnRetryTimes, err.Error())
		},
		MaxIdleConns:        1000,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     10 * time.Second,
	}

	return &httpClient{
		client:  http.Client{Transport: transport},
		addrMan: addrMan,
		cluster: cluster,
	}, nil
}

func (c *httpClient) do(req *http.Request) (*http.Response, error) {
	if c.cluster != "" && !isIPAddr(req.URL.Host) {
		// add cluster name as a part of host
		//	in order to identity http connection pool
		req.URL.Host += clusterSep + c.cluster
	}
	return c.client.Do(req)
}
