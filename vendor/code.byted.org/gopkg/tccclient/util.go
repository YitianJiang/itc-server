package tccclient

import (
	"code.byted.org/gopkg/env"
	"code.byted.org/gopkg/logs"
)

func checkClientConfig(client *Client) {
	psm := env.PSM()
	if psm != client.serviceName {
		logs.Warn("TCC check: service name does not match that gets from gopkg/env: %s", psm)
	}

	cluster := env.Cluster()
	if cluster != client.cluster {
		logs.Info("TCC check: cluster you specified does not match that gets from gopkg/env: %s. But it's ok to just use the default", cluster)
	}

	product := env.IsProduct()
	if !product {
		logs.Warn("TCC check: env does not match that gets from gopkg/env, isProduct: %v", product)
	}
}
