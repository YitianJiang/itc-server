package hystrix

import (
	"math"
	"sync"
	"time"
)

const (
	// DefaultTimeout is how long to wait for command to complete, in milliseconds
	DefaultTimeout = 1000
	// DefaultMaxConcurrent is how many commands of the same type can run at the same time
	DefaultMaxConcurrent = 10
	// DefaultVolumeThreshold is the minimum number of requests needed before a circuit can be tripped due to health
	DefaultVolumeThreshold = 20
	// DefaultSleepWindow is how long, in milliseconds, to wait after a circuit opens before testing for recovery
	DefaultSleepWindow = 5000
	// DefaultErrorPercentThreshold causes circuits to open once the rolling measure of errors exceeds this percent of requests
	DefaultErrorPercentThreshold = 50
	// DefaultCircuitBreakerEnabled is to enable circuit breaker
	DefaultCircuitBreakerEnabled = false
)

type settings struct {
	Timeout                time.Duration
	MaxConcurrentRequests  int
	RequestVolumeThreshold uint64
	SleepWindow            time.Duration
	ErrorPercentThreshold  int
	CircuitBreakerEnabled  bool
}

// CommandConfig is used to tune circuit settings at runtime
type CommandConfig struct {
	Timeout                int  `json:"timeout" yaml:"timeout"`
	MaxConcurrentRequests  int  `json:"max_concurrent_requests" yaml:"max_concurrent_requests"`
	RequestVolumeThreshold int  `json:"request_volume_threshold" yaml:"request_volume_threshold"`
	SleepWindow            int  `json:"sleep_window" yaml:"sleep_window"`
	ErrorPercentThreshold  int  `json:"error_percent_threshold" yaml:"error_percent_threshold"`
	CircuitBreakerEnabled  bool `json:"circuit_breaker_enabled" yaml:"circuit_breaker_enabled"`
}

var circuitSettings map[string]*settings
var settingsMutex *sync.RWMutex

func init() {
	circuitSettings = make(map[string]*settings)
	settingsMutex = &sync.RWMutex{}
}

// Configure applies settings for a set of circuits
func Configure(cmds map[string]CommandConfig) {
	for k, v := range cmds {
		ConfigureCommand(k, v)
	}
}

var noOpSettings = &settings{
	Timeout:                time.Duration(math.MaxInt64),
	MaxConcurrentRequests:  100000, // We think there is no chance volume exceed 100000 in a single host
	RequestVolumeThreshold: 100000,
	SleepWindow:            0,
	ErrorPercentThreshold:  101,
	CircuitBreakerEnabled:  false,
}

// ConfigureCommand applies settings for a circuit
func ConfigureCommand(name string, config CommandConfig) {
	settingsMutex.Lock()
	defer settingsMutex.Unlock()

	circuitBreakersEnabled := DefaultCircuitBreakerEnabled
	if config.CircuitBreakerEnabled {
		circuitBreakersEnabled = config.CircuitBreakerEnabled
	}

	if !circuitBreakersEnabled {
		circuitSettings[name] = noOpSettings
		return
	}

	timeout := DefaultTimeout
	if config.Timeout != 0 {
		timeout = config.Timeout
	}

	max := DefaultMaxConcurrent
	if config.MaxConcurrentRequests != 0 {
		max = config.MaxConcurrentRequests
	}

	volume := DefaultVolumeThreshold
	if config.RequestVolumeThreshold != 0 {
		volume = config.RequestVolumeThreshold
	}

	sleep := DefaultSleepWindow
	if config.SleepWindow != 0 {
		sleep = config.SleepWindow
	}

	errorPercent := DefaultErrorPercentThreshold
	if config.ErrorPercentThreshold != 0 {
		errorPercent = config.ErrorPercentThreshold
	}

	circuitSettings[name] = &settings{
		Timeout:                time.Duration(timeout) * time.Millisecond,
		MaxConcurrentRequests:  max,
		RequestVolumeThreshold: uint64(volume),
		SleepWindow:            time.Duration(sleep) * time.Millisecond,
		ErrorPercentThreshold:  errorPercent,
		CircuitBreakerEnabled:  circuitBreakersEnabled,
	}
}

func getSettings(name string) *settings {
	settingsMutex.RLock()
	s, exists := circuitSettings[name]
	settingsMutex.RUnlock()

	if !exists {
		ConfigureCommand(name, CommandConfig{})
		s = getSettings(name)
	}

	return s
}
