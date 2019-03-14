package connpool

import (
	"net"
	"time"
)

// Dialer .
type Dialer interface {
	Dial(addr string, timeout time.Duration) (net.Conn, error)
}

type innerDialer struct{}

func (td *innerDialer) Dial(addr string, timeout time.Duration) (net.Conn, error) {
	d := net.Dialer{Timeout: timeout}
	// FIXME: replace Dialer with net.DialXX
	n := len(addr)
	if addr[n-1:n] == ":" {
		return d.Dial("unix", addr[:n-1])
	} else {
		return d.Dial("tcp", addr)
	}
}

// ConnPool .
type ConnPool interface {
	// Get returns a new connection from the pool;
	// Closing the connections puts it back to the Pool for longconn;
	Get(host, port string, connTimeout time.Duration) (net.Conn, error)
}

type LongConnPool interface {
	ConnPool

	// Clean the state maintained in the poor for this address
	Clean(host, port string)
}
