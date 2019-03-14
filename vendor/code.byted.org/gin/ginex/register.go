package ginex

import (
	"fmt"
	"os"

	"code.byted.org/golf/consul"
	"code.byted.org/gopkg/logs"
)

const (
	IS_LOAD_REGISTERED = "IS_LOAD_REGISTERED"
)

var (
	register *consul.RegisterContext
)

// Register registers service address and useful meta data to consul.
// It's will give up register if:
//   - service has been registered by load.sh
//   - it's not running in product mode
func Register() (err error) {
	if os.Getenv(IS_LOAD_REGISTERED) == "1" {
		logs.Info("Skip self-register: Load has registered")
		return nil
	}
	if !Product() {
		logs.Warn("Skip self-register: not in product environment")
		return nil
	}

	logs.Info("Register service: %s, cluster:%s", PSM(), Cluster())
	register, err = consul.InitRegister()
	if err != nil {
		logs.Errorf("Failed to init register: %s", err)
		return err
	}
	register.DefineService(PSM(), appConfig.ServicePort, map[string]string{
		"version": appConfig.ServiceVersion,
		"cluster": Cluster(),
	}, -1)
	err = register.StartRegister()
	if err != nil {
		logs.Errorf("Failed to start register: %s", err)
	} else {
		logs.Infof("Successfully start register: %s", PSM())
	}
	return
}

// StopRegister stops register loop and deregisters service
func StopRegister() (err error) {
	if register == nil {
		return nil
	}
	if err = register.StopRegister(); err != nil {
		return fmt.Errorf("Failed to stop register: %s", err.Error())
	}
	serviceId := fmt.Sprintf("%s-%d", PSM(), appConfig.ServicePort)
	agent := register.Sd.Client.Agent()
	if err = agent.ServiceDeregister(serviceId); err != nil {
		return fmt.Errorf("Failed to deregister service: %s, %s", serviceId, err)
	} else {
		logs.Infof("Deregister service: %s", serviceId)
	}
	return nil
}
