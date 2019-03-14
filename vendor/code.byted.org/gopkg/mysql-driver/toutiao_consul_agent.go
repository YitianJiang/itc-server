package mysql

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"os"
	"time"
)

var consulAddr string

func init() {
	host := "127.0.0.1"
	port := "2280"
	if s := os.Getenv("CONSUL_HTTP_HOST"); s != "" {
		host = s
	}
	if s := os.Getenv("CONSUL_HTTP_PORT"); s != "" {
		port = s
	}
	if s := os.Getenv("TCE_HOST_IP"); s != "" {
		host = s
		port = "2280"
	}
	if s1, s2, _ := net.SplitHostPort(os.Getenv("CONSUL_HTTP_ADDR")); s1 != "" && s2 != "" {
		host = s1
		port = s2
	}
	consulAddr = host + ":" + port
}

type ConsulEndpoint struct {
	Host string
	Port int
	Tags map[string]string
}

func consulGet(service string) ([]ConsulEndpoint, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond*500)
	defer cancel()
	req, err := http.NewRequest("GET", "http://"+consulAddr+"/v1/lookup/name?name="+service, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	ret := make([]ConsulEndpoint, 0, 50)
	return ret, json.NewDecoder(resp.Body).Decode(&ret)
}
