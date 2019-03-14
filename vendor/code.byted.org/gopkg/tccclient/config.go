package tccclient

const (
	DefaultCluster = "default"
	DefaultEnv     = "prod"
)

type Config struct {
	// ServiceName string
	Cluster string
	Env     string

	DisableMetrics bool
}

func NewConfig() *Config {
	c := Config{
		Cluster: DefaultCluster,
		Env:     DefaultEnv,
	}
	return &c
}
