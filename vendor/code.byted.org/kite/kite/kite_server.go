/*
 *  KITE RPC FRAMEWORK
 */

package kite

import (
	"errors"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"syscall"
	"time"

	"code.byted.org/gopkg/env"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/thrift"
	"code.byted.org/kite/kite/gls"
	"code.byted.org/kite/kitutil/kvstore"
)

// server metrics
const (
	errorServerThroughputFmt string = "service.thrift.%s.calledby.error.throughput"
)

type RpcServer struct {
	l net.Listener

	processorFactory thrift.TProcessorFactory
	transportFactory thrift.TTransportFactory
	protocolFactory  thrift.TProtocolFactory

	remoteConfiger *remoteConfiger
	overloader     *overloader
}

// NewRpcServer create the global server instance
func NewRpcServer() *RpcServer {
	// Using buffered transport and binary protocol as default,
	// buffer size is 4096
	var s *RpcServer
	if ServiceMeshMode {
		oriTransport := thrift.NewTBufferedTransportFactory(DefaultTransportBufferedSize)
		transport := thrift.NewHeaderTransportFactory(oriTransport)
		protocol := thrift.NewHeaderProtocolFactory()
		s = &RpcServer{
			transportFactory: transport,
			protocolFactory:  protocol,
		}
	} else {
		transport := thrift.NewTBufferedTransportFactory(DefaultTransportBufferedSize)
		protocol := thrift.NewTBinaryProtocolFactoryDefault()
		s = &RpcServer{
			transportFactory: transport,
			protocolFactory:  protocol,
		}
	}

	// FIXME:when mesh mode, we don't need remoteConfiger&overloader,
	// but there are more functions depend with those,
	// so replace a empty kvstore with ETCD
	if ServiceMeshMode {
		s.remoteConfiger = newRemoteConfiger(s, newEmptyStorer())
		s.overloader = newOverloader(limitMaxConns, limitQPS)
	} else {
		s.remoteConfiger = newRemoteConfiger(s, kvstore.NewETCDStorer())
		s.overloader = newOverloader(limitMaxConns, limitQPS)
	}
	return s
}

// ListenAndServe ...
func (p *RpcServer) ListenAndServe() error {
	var addr string
	if ServiceMeshMode {
		addr = ServiceMeshIngressAddr
		if _, err := net.ResolveIPAddr("tcp", addr); err == nil {
			ListenType = LISTEN_TYPE_TCP
		} else {
			ListenType = LISTEN_TYPE_UNIX
			syscall.Unlink(addr)
		}
	} else {
		if ListenType == LISTEN_TYPE_TCP {
			addr = ServiceAddr + ServicePort
		} else if ListenType == LISTEN_TYPE_UNIX {
			addr = ServiceAddr
			syscall.Unlink(ServiceAddr)
		} else {
			return errors.New(fmt.Sprintf("Invalid listen type %s", ListenType))
		}
	}

	l, err := net.Listen(ListenType, addr)
	if err != nil {
		return err
	}
	if ListenType == LISTEN_TYPE_UNIX {
		os.Chmod(ServiceAddr, os.ModePerm)
	}
	return p.Serve(l)
}

// Serve ...
func (p *RpcServer) Serve(ln net.Listener) error {
	if p.l != nil {
		panic("KITE: Listener not nil")
	}
	p.l = ln
	logs.Info("KITE: server listening on %s", ln.Addr())

	if Processor == nil {
		panic("KITE: Processor is nil")
	}
	p.processorFactory = thrift.NewTProcessorFactory(Processor)

	p.startDaemonRoutines()

	for {
		// If l.Close() is called will return closed error
		conn, err := p.l.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "closed") {
				return err
			}

			logs.Errorf("KITE: accept failed, err:%v", err)

			time.Sleep(10 * time.Millisecond) // too many open files ?
			continue
		}

		if !ServiceMeshMode {
			if !p.overloader.TakeConn() {
				msg := fmt.Sprintf("KITE: connection overload, limit=%v, now=%v, remote=%s\n",
					p.overloader.ConnLimit(), p.overloader.ConnNow(), conn.RemoteAddr().String())
				logs.Warnf(msg)
				p.onConnOverload()
				conn.Close()
				continue
			}
			if !p.overloader.TakeQPS() {
				msg := fmt.Sprintf("KITE: qps overload, qps_limit=%v, remote=%s\n", p.overloader.QPSLimit(), conn.RemoteAddr().String())
				logs.Warnf(msg)
				conn.Close()
				p.onQPSOverload()
				p.overloader.ReleaseConn()
				continue
			}
		}

		go func(conn net.Conn) {
			if GetRealIP { // encode real ip to stack
				handleRPC := func() {
					client := thrift.NewTSocketFromConnTimeout(conn, ReadWriteTimeout)
					if err := p.processRequests(client); err != nil {
						logs.Warnf("KITE: processing request error=%s, remote=%s", err, conn.RemoteAddr().String())
					}
				}

				addrCode := encodeAddr(conn.RemoteAddr().String())
				if addrCode == 0 {
					logs.Debugf("KITE: invalid remote ip: %s", conn.RemoteAddr().String())
					handleRPC()
				} else {
					gls.SetGID(addrCode, handleRPC)
				}
			} else {
				client := thrift.NewTSocketFromConnTimeout(conn, ReadWriteTimeout)
				if err := p.processRequests(client); err != nil {
					logs.Warnf("KITE: processing request error=%s, remote=%s", err, conn.RemoteAddr().String())
				}
			}
		}(conn)
	}
}

// Stop .
func (p *RpcServer) Stop() error {
	if p.l == nil {
		return nil
	}
	if err := p.l.Close(); err != nil {
		return err
	}
	deadline := time.After(ExitWaitTime)
	for {
		select {
		case <-deadline:
			return errors.New("deadline excceded")
		default:
			if p.overloader.ConnNow() == 0 {
				return nil
			}
			time.Sleep(time.Millisecond)
		}
	}
}

func (p *RpcServer) processRequests(client thrift.TTransport) error {
	processor := p.processorFactory.GetProcessor(client)
	transport := p.transportFactory.GetTransport(client)
	protocol := p.protocolFactory.GetProtocol(transport)

	defer func() {
		if e := recover(); e != nil {
			if err, ok := e.(string); !ok || err != recoverMW {
				const size = 64 << 10
				buf := make([]byte, size)
				buf = buf[:runtime.Stack(buf, false)]
				logs.Fatal("KITE: panic in processor: %s: %s", e, buf)
			}
			p.onPanic()
		}
		p.overloader.ReleaseConn()
	}()
	defer transport.Close()
	// This loop for processing request on a connection.
	var count = 0
	for {
		count++
		metricsClient.EmitCounter("kite.process.throughput", 1, "", map[string]string{"name": ServiceName, "cluster": env.Cluster()})
		ok, err := processor.Process(protocol, protocol)
		if err, ok := err.(thrift.TTransportException); ok {
			if err.TypeId() == thrift.END_OF_FILE ||
				// TODO(xiangchao.01): this timeout maybe not precision,
				// fix should in thrift package later.
				err.TypeId() == thrift.TIMED_OUT {
				return nil
			}
			if err.TypeId() == thrift.UNKNOWN_METHOD {
				name := fmt.Sprintf("toutiao.service.thrift.%s.process.error", ServiceName)
				metricsClient.EmitCounter(name, 1, "", map[string]string{
					"name":    "UNKNOWN_METHOD",
					"cluster": env.Cluster(),
				})
			}
		}

		if err != nil {
			return err
		}
		if !ok {
			break
		}
		// 当请求是短连接的时候，会在第二次循环的时候，读取到thrift.END_OF_FILE的错误
		// 为了兼容这种情况，当count等于1的时候，不应该执行onProcess调用
		if count == 1 {
			continue
		}
		if !p.overloader.TakeQPS() {
			msg := "KITE: qps overload, close socket forcely"
			logs.Warnf(msg)
			return errors.New("KITE: qps overload, close socket forcely")
		}
	}
	return nil
}

func (p *RpcServer) getRPCConfig(r RPCMeta) RPCConfig {
	c, err := p.remoteConfiger.GetRemoteRPCConfig(r)
	if err != nil {
		logs.Warnf("KITE: get remote config for %v err: %s, default config will be used", r, err.Error())
		return defaultRPCConfig
	}
	return c
}

func (p *RpcServer) onPanic() {
	metrics := fmt.Sprintf("service.thrift.%s.panic", ServiceName)
	metricsClient.EmitCounter(metrics, 1, "", map[string]string{"name": ServiceName, "cluster": env.Cluster()})
}

// RemoteConfigs .
func (p *RpcServer) RemoteConfigs() map[string]interface{} {
	return p.remoteConfiger.AllRemoteConfigs()
}

// Overload .
func (p *RpcServer) Overload() (connNow, connLim, qpsLim int64) {
	return p.overloader.ConnNow(), p.overloader.ConnLimit(), p.overloader.QPSLimit()
}

func (p *RpcServer) onListenFailed() {
	p.serverErrorMetrics("listen_failed")
}

func (p *RpcServer) onAcceptFailed() {
	p.serverErrorMetrics("accept_failed")
}

func (p *RpcServer) onConnOverload() {
	p.serverErrorMetrics("conn_overload")
}

func (p *RpcServer) onQPSOverload() {
	p.serverErrorMetrics("qps_overload")
}

func (p *RpcServer) serverErrorMetrics(errorType string) {
	metrics := fmt.Sprintf(errorServerThroughputFmt, ServiceName)
	metricsClient.EmitCounter(metrics, 1, "", map[string]string{"cluster": env.Cluster(), "type": errorType})
}
