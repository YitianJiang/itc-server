package connpool

import (
	"fmt"
	"net"
	"time"
)

// ShortPool .
type ShortPool struct {
	Dialer    Dialer // hack this for test
	targetPSM string
}

// NewShortPool timeout is connection timeout.
func NewShortPool(targetPSM string) *ShortPool {
	return &ShortPool{
		Dialer:    &innerDialer{},
		targetPSM: targetPSM,
	}
}

// Get return a PoolConn instance which implemnt net.Conn interface.
func (p *ShortPool) Get(host, port string, timeout time.Duration) (net.Conn, error) {
	addr := host + ":" + port
	conn, err := p.Dialer.Dial(addr, timeout)
	if err != nil {
		getShortConnFail(p.targetPSM, addr)
		return nil, fmt.Errorf("dial connection err: %s, addr: %s", err, addr)
	}
	getShortConnSucc(p.targetPSM, addr)
	return conn, nil
}
