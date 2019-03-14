package consul

import (
	"log"
	"math"
	"os"
	"time"
)

type ServiceInfo struct {
	Name string
	Port int
	Tags map[string]string
	TTL  int
}

type RegisterContext struct {
	Sd          *ServiceDiscovery
	services    []*ServiceInfo
	registering bool
	ensure_safe bool
	logger      *log.Logger
}

func InitRegister() (*RegisterContext, error) {
	sd, err := NewServiceDiscovery()
	if err != nil {
		return nil, err
	}
	return &RegisterContext{sd, nil, false, true, log.New(os.Stderr, "consul", 0)}, nil
}

func (ctx *RegisterContext) WithoutEnsureSafe() {
	ctx.ensure_safe = false
}

func (ctx *RegisterContext) DefineService(name string, port int, tags map[string]string, ttl int) error {
	if ttl <= 0 {
		ttl = 120
	}
	name = addPerfPrefix(name)
	ctx.services = append(ctx.services, &ServiceInfo{name, port, tags, ttl})
	ctx.logger.Printf("Defined service %s, port %d, ttl %d\n", name, port, ttl)
	return nil
}

func (ctx *RegisterContext) registerService(service *ServiceInfo) {
	attempt := 0
	for {
		if !ctx.registering {
			return
		}
		next_lease, err := ctx.Sd.Announce(service.Name, service.Port, service.Tags, service.TTL)
		if err != nil {
			// bounded exponential backoff
			next_lease = int(math.Min(0.2*math.Pow(2, float64(attempt)), float64(service.TTL)*0.9))
			attempt++
		} else {
			attempt = 0
		}
		time.Sleep(time.Duration(int(math.Max(float64(next_lease), 1))) * time.Second)
	}
}

func (ctx *RegisterContext) StartRegister() error {
	ctx.registering = true
	ctx.logger.Printf("Starting to register defined services\n")
	for _, svc := range ctx.services {
		if ctx.ensure_safe {
			if safe, err := ensureSafety(ctx.Sd, svc.Name); !safe {
				ctx.logger.Printf("Not safe to register %s, reason: %s, skipping\n", svc.Name, err.Error())
				continue
			}
		}
		go ctx.registerService(svc)
	}
	return nil
}

func (ctx *RegisterContext) StopRegister() error {
	ctx.logger.Printf("Stopping to register defined services\n")
	ctx.registering = false
	return nil
}
