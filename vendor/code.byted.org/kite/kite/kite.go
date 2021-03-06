package kite

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"code.byted.org/gopkg/env"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/thrift"
	"code.byted.org/kite/endpoint"
	"code.byted.org/kite/kitc"
	trace "code.byted.org/trace/trace-client-go"
	opentracing "github.com/opentracing/opentracing-go"
)

const (
	KiteVersion = "v3.2.2"
)

var (
	// RPCServer is singleton RpcServer instance
	RPCServer *RpcServer

	// Processor is thrift Processor, set by code generated by kitool
	Processor thrift.TProcessor

	// IDLs this service use
	IDLs = make(map[string]string) // map[filename]content
)

// Init .
func Init() {
	// in tce env, MY_CPU_LIMIT will be set as limit cpu cores.
	if v := os.Getenv("GOMAXPROCS"); v == "" {
		if v := os.Getenv("MY_CPU_LIMIT"); v != "" {
			n, err := strconv.ParseInt(v, 10, 64)
			if err == nil {
				runtime.GOMAXPROCS(int(n))
			}
		}
	}

	// conf priority: file > envs > args
	initFromArgs()
	initFromEnvs()
	initFromConfFile()
	if ConfigFile == "" {
		fmt.Fprintf(os.Stderr, "configfile is empty, use -conf option or %s environment", _ENV_CONFIG_FILE)
		Usage()
	}
	if LogDir == "" {
		fmt.Fprintf(os.Stderr, "logdir is empty, use -log option or %s environment", _ENV_LOG_DIR)
		Usage()
	}
	if ServiceName == "" {
		fmt.Fprintf(os.Stderr, "servicename is empty, use -svc option or %s environment", _ENV_SERVICE_NAME)
		Usage()
	}
	if ServicePort == "" {
		fmt.Fprintln(os.Stderr, "no service port, please checkout your args and conffile")
		os.Exit(-1)
	}

	// init databus config
	initDatabusChannel()
	if !EnableMetrics {
		metricsClient = &EmptyEmiter{}
	}

	// init tracing component
	if EnableTracing {
		if err := trace.Init(ServiceName); err != nil {
			opentracing.SetGlobalTracer(opentracing.NoopTracer{})
		}
	}

	// init access logger
	filename := filepath.Join(filepath.Join(LogDir, "app"), ServiceName+".access.log")
	AccessLogger = newRPCLogger(filename)
	fmt.Printf("KITE: access log path: %s\n", filename)

	// init call logger for kitc
	filename = filepath.Join(filepath.Join(LogDir, "rpc"), ServiceName+".call.log")
	CallLogger = newRPCLogger(filename)
	fmt.Printf("KITE: call log path: %s\n", filename)
	kitc.SetCallLog(CallLogger)

	if FileLog {
		// note: we use hourly roated log for the MSP service depend on this
		initFileProvider(LogLevel, LogFile, logs.HourDur, MaxLogSize)
	}
	if ConsoleLog {
		initConsoleProvider()
	}
	initDatabusProvider(LogLevel)
	initAgentProvider(LogLevel)
	initDefaultLog(LogLevel)

	if ServiceMeshMode {
		fmt.Fprintf(os.Stdout, "KITE: open service mesh ingress mode at %s\n", ServiceMeshIngressAddr)
	}

	// create the singleton server instance
	RPCServer = NewRpcServer()
}

// Run starts the default RpcServer.
// It blocks until recv SIGTERM, SIGHUP or SIGINT.
func Run() error {
	errCh := make(chan error, 1)
	go func() { errCh <- RPCServer.ListenAndServe() }()
	if err := waitSignal(errCh); err != nil {
		return err
	}
	return RPCServer.Stop()
}

// RunWithListener starts the default RpcServer with the Listener ln.
// It's similar to Run
func RunWithListener(ln net.Listener) error {
	if ServiceMeshMode && ServiceMeshIngressAddr != "" && ServiceMeshIngressAddr != ln.Addr().String() {
		logs.Warnf("KITE: using service mesh: expected listening on %s actual %s", ServiceMeshIngressAddr, ln.Addr())
	}
	errCh := make(chan error, 1)
	go func() { errCh <- RPCServer.Serve(ln) }()
	if err := waitSignal(errCh); err != nil {
		return err
	}
	return RPCServer.Stop()
}

func waitSignal(errCh chan error) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	// flush opentracing spans in queue when server exit
	defer trace.Close()

	// service may not be available as soon as startup,  delay registeration to consul.
	// It doesn't work in TCE env.
	delayRegister := time.After(10 * time.Second)

	for {
		select {
		case sig := <-signals:
			switch sig {
			case syscall.SIGTERM:
				return errors.New(sig.String()) // force exit
			case syscall.SIGHUP, syscall.SIGINT:
				return nil // graceful shutdown
			}
		case err := <-errCh:
			return err
		case <-delayRegister:

			// Auto register to consul.
			// It doesn't work in TCE env.
			if err := Register(); err != nil {
				logs.Errorf("KITE: Register service error: %s", err)
			}
			defer StopRegister()

		}
	}
}

// MethodContext called by code generated by kitool to get a inited context
func MethodContext(method string) context.Context {
	// ??????????????????????????????, ?????????????????????, ???????????????????????????????????????????????????;
	// ???????????????????????????????????????????????????, ?????????????????????????????????????????????;
	rpcInfo := &RPCInfo{
		RPCMeta: RPCMeta{
			Method: method,
		},
	}

	return newCtxWithRPCInfo(context.Background(), rpcInfo)
}

// DefineProcessor called by code generated by kitool to set Processer
func DefineProcessor(p thrift.TProcessor) {
	if Processor != nil {
		panic("DefineProcessor more than onece")
	}
	Processor = p
}

// SetIDL called by code generated by kitool to define method
func SetIDL(filename, content string) {
	IDLs[filename] = content
}

// KiteMW wrap every endpoint in this service, called by code generated by kitool
func KiteMW(next endpoint.EndPoint) endpoint.EndPoint {
	var mids, optMids []endpoint.Middleware

	if EnableTracing {
		mids = append(mids, OpenTracingMW)
	}
	if !ServiceMeshMode {
		optMids = []endpoint.Middleware{
			ExposeCtxMW,
			AccessLogMW,
			AccessMetricsMW,
			ACLMW,
			EndpointQPSLimitMW,
			StressBotMW,
			RecoverMW,
			AdditionMW,
			BaseRespCheckMW,
		}
	} else {
		optMids = []endpoint.Middleware{
			ExposeCtxMW,
			AccessLogMW,
			AccessMetricsMW,
			RecoverMW,
			AdditionMW,
			BaseRespCheckMW,
		}
	}
	mids = append(mids, optMids...)
	if !NoPushNotice {
		mids = append(mids, PushNoticeMW)
	}
	mid := endpoint.Chain(ParserMW, mids...)
	return mid(next)
}

var mMap = make(map[string]endpoint.Middleware)
var userMW endpoint.Middleware

// AddMethodMW use a middleware for a define method.
func AddMethodMW(m string, mws ...endpoint.Middleware) {
	if len(mws) >= 1 {
		mMap[m] = endpoint.Chain(mws[0], mws[1:]...)
	}
}

// Use middlewares will enable for all this service's method.
func Use(mws ...endpoint.Middleware) {
	if len(mws) >= 1 {
		if userMW != nil {
			userMW = endpoint.Chain(userMW, mws...)
		} else {
			userMW = endpoint.Chain(mws[0], mws[1:]...)
		}
	}
}

// GetLocalIp for compatibility
func GetLocalIp() string {
	return env.HostIP()
}

// NewCallLogger for compatibility
func NewCallLogger(logDir string, serviceName string, useScribe bool) *logs.Logger {
	filename := filepath.Join(logDir, serviceName+".call.log")
	return newRPCLogger(filename)
}

// NewAccessLogger for compatibility
func NewAccessLogger(logDir string, serviceName string, useScribe bool) *logs.Logger {
	filename := filepath.Join(logDir, serviceName+".access.log")
	return newRPCLogger(filename)
}
