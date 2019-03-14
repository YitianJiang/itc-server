package kite

import (
	"fmt"
	"os"
	"strconv"

	"code.byted.org/golf/consul"
	"code.byted.org/gopkg/env"
)

var (
	register *consul.RegisterContext
)

// Register write its name into consul for other services lookup
func Register() error {
	if !env.IsProduct() && os.Getenv("IS_PROD_RUNTIME") == "" {
		// Only register in prod or IS_PROD_RUNTIME is setted
		return nil
	}
	if os.Getenv("IS_LOAD_REGISTERED") == "1" {
		// load script has registed
		return nil
	}

	if ListenType != LISTEN_TYPE_TCP {
		return nil
	}

	var err error
	register, err = consul.InitRegister()
	if err != nil {
		return err
	}
	port, err := strconv.Atoi(ServicePort)
	if err != nil {
		return fmt.Errorf("parse service port %s", err)
	}
	tags := map[string]string{
		"transport": "thrift.TBufferedTransport",
		"protocol":  "thrift.TBinaryProtocol",
		"version":   ServiceVersion,
		"cluster":   ServiceCluster,
	}
	if ServiceShard != "" {
		tags["shard"] = ServiceShard
	}

	register.DefineService(ServiceName, port, tags, -1)
	return register.StartRegister()
}

// StopRegister stops register loop and deregisters service
func StopRegister() error {
	if register == nil {
		return nil
	}
	if err := register.StopRegister(); err != nil {
		return fmt.Errorf("Failed to stop register: %s", err.Error())
	}
	serviceID := fmt.Sprintf("%s-%s", ServiceName, ServicePort)
	agent := register.Sd.Client.Agent()
	if err := agent.ServiceDeregister(serviceID); err != nil {
		return fmt.Errorf("Failed to deregister service: %s, %s", serviceID, err)
	}
	return nil
}
