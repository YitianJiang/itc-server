package consul

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	consulUnixSock = "/opt/tmp/sock/consul.sock"
	cacheTime      = 15 * time.Second
)

var nowfunc = time.Now

var consul *Consul
var dialer = net.Dialer{Timeout: 3 * time.Second}

type lookupopts struct {
	idc     string
	cluster string
	limit   int
	nocache bool
}

type LookupOptions func(oo *lookupopts)

type IDC string

//IDC
const (
	IDCLF     IDC = "lf"
	IDCHL     IDC = "hl"
	IDCMALIVA IDC = "maliva"
	IDCALISG  IDC = "alisg"
)

func WithLimit(n int) LookupOptions {
	return func(oo *lookupopts) {
		oo.limit = n
	}
}

// WithIDC tells client to fetch instances
//	from specified idc instead of local idc
func WithIDC(idc IDC) LookupOptions {
	return func(oo *lookupopts) {
		oo.idc = string(idc)
	}
}

func WithCluster(cluster string) LookupOptions {
	return func(oo *lookupopts) {
		oo.cluster = cluster
	}
}

func WithNoCache(nocache bool) LookupOptions {
	return func(oo *lookupopts) {
		oo.nocache = nocache
	}
}

func init() {
	host := ""
	for _, name := range []string{"CONSUL_HTTP_HOST", "MY_HOST_IP", "TCE_HOST_IP"} {
		if v := os.Getenv(name); v != "" {
			host = v
			break
		}
	}
	if host == "" {
		if _, err := os.Stat(consulUnixSock); err == nil {
			consul = NewConsul(consulUnixSock)
			return
		}
		host = "127.0.0.1"
	}
	defaultPort := "2280"
	if port := os.Getenv("CONSUL_HTTP_PORT"); port != "" {
		defaultPort = port
	}
	consul = NewConsul(fmt.Sprintf("%s:%s", host, defaultPort))
}

func DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	return consul.DialContext(ctx, network, addr)
}

func Dialer(name string, timeout time.Duration,
	opts ...LookupOptions) func(ctx context.Context) (net.Conn, error) {
	return consul.Dialer(name, timeout, opts...)
}

func Lookup(name string, opts ...LookupOptions) (Endpoints, error) {
	return consul.Lookup(name, opts...)
}

type cachedEndpoints struct {
	UpdatedAt time.Time
	Endpoints Endpoints
}

type Consul struct {
	cli *http.Client

	mu sync.RWMutex
	m  map[string]*cachedEndpoints

	nlookupraw int64
}

func NewConsul(addr string) *Consul {
	consul := &Consul{m: make(map[string]*cachedEndpoints)}
	consul.cli = &http.Client{Timeout: 500 * time.Millisecond}
	if strings.HasPrefix(addr, "/") { // unix sock?
		consul.cli.Transport = &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return dialer.DialContext(ctx, "unix", consulUnixSock)
			},
		}
	} else {
		consul.cli.Transport = &http.Transport{
			DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
				return dialer.DialContext(ctx, "tcp", addr)
			},
		}
	}
	return consul
}

func (c *Consul) Lookup(name string, opts ...LookupOptions) (Endpoints, error) {
	t := nowfunc()

	oo := lookupopts{}
	for _, op := range opts {
		op(&oo)
	}

	key := name + "|" + oo.cluster + "|" + oo.idc

	c.mu.RLock()
	e := c.m[key]
	c.mu.RUnlock()
	if e != nil && t.Sub(e.UpdatedAt) < cacheTime && !oo.nocache {
		return e.Endpoints, nil
	}

	defer func() {
		c.mu.Lock()
		if e != nil {
			e.UpdatedAt = t
			c.m[key] = e
		}
		c.mu.Unlock()
	}()

	ee, err := c.lookup(name, oo)
	if err != nil {
		if oo.nocache {
			return nil, err
		}
		if e != nil {
			log.Printf("gopkg/consul: %s", err)
			return e.Endpoints, nil
		}
		return nil, err
	}
	ret := make(Endpoints, len(ee))
	for i, e := range ee {
		ret[i] = e.parse()
	}
	if oo.cluster != "" {
		ret = ret.FilterCluster(oo.cluster)
	}
	e = &cachedEndpoints{
		Endpoints: ret,
	}
	return ret, nil
}

func (c *Consul) lookup(name string, oo lookupopts) ([]ConsulEndpoint, error) {
	atomic.AddInt64(&c.nlookupraw, 1)
	uv := url.Values{}
	if !strings.Contains(name, ".service.") && oo.idc != "" {
		name += ".service." + oo.idc
	}
	uv.Set("name", name)
	if oo.limit > 0 {
		uv.Set("limit", strconv.Itoa(oo.limit))
	}
	if oo.cluster != "" {
		uv.Set("cluster", oo.cluster)
	}
	url := "http://127.0.0.1:2280/v1/lookup/name?" + uv.Encode()
	resp, err := c.cli.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, errors.New(string(b))
	}
	var ee []ConsulEndpoint
	if err := json.NewDecoder(resp.Body).Decode(&ee); err != nil {
		return nil, err
	}
	return ee, nil
}

func (c *Consul) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	s, _, err := net.SplitHostPort(addr)
	if err != nil || net.ParseIP(s) != nil { // if host is IP
		return dialer.DialContext(ctx, network, addr)
	}
	ee, err := c.Lookup(s)
	if err != nil {
		return nil, err
	}
	return dialer.DialContext(ctx, network, ee.GetOne().Addr)
}

func (c *Consul) Dialer(name string, timeout time.Duration,
	opts ...LookupOptions) func(ctx context.Context) (net.Conn, error) {

	return func(ctx context.Context) (net.Conn, error) {
		for i := 0; ; i++ {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			_timeout := timeout
			if t, ok := ctx.Deadline(); ok {
				if d := t.Sub(nowfunc()); d < _timeout {
					_timeout = d
				}
			}
			ee, err := c.Lookup(name, opts...)
			if err != nil {
				return nil, err
			}
			if len(ee) == 0 {
				return nil, errors.New("no endpoint")
			}
			conn, err := net.DialTimeout("tcp", ee.GetOne().Addr, _timeout)
			if err == nil || i >= 3 {
				return conn, err
			}
		}
	}
}

type ServiceDefinition struct {
	ID    string    // name-port
	Name  string    // service name
	Port  int       // service port
	Tags  []string  `json:",omitempty"`
	Check CheckType `json:",omitempty"`
}

type CheckType struct {
	Interval time.Duration `json:",omitempty"`
	Script   string        `json:",omitempty"`
}

func (c *Consul) Register(s ServiceDefinition) error {
	if s.ID == "" {
		s.ID = fmt.Sprintf("%s-%d", s.Name, s.Port)
	}
	b, _ := json.Marshal(s)
	url := "http://127.0.0.1:2280/v1/agent/service/register"
	resp, err := c.cli.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return nil
	}
	b, _ = ioutil.ReadAll(resp.Body)
	return errors.New(string(b))
}

func (c *Consul) Deregister(name string, port int) error {
	url := fmt.Sprintf("http://127.0.0.1:2280/v1/agent/service/deregister/%s-%d", name, port)
	resp, err := c.cli.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return nil
	}
	b, _ := ioutil.ReadAll(resp.Body)
	return errors.New(string(b))
}
