package connpool

import (
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type longConn struct {
	net.Conn
	peer     *peer
	err      error
	deadline time.Time
	sync.RWMutex
}

func (c *longConn) Close() error {
	return c.peer.put(c)
}

func (c *longConn) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	if err != nil {
		c.Lock()
		c.err = err
		c.Unlock()
	}
	return n, err
}

func (c *longConn) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	if err != nil {
		c.Lock()
		c.err = err
		c.Unlock()
	}
	return n, err
}

func (c *longConn) SetDeadline(t time.Time) error {
	err := c.Conn.SetDeadline(t)
	if err != nil {
		c.Lock()
		c.err = err
		c.Unlock()
	}
	return err
}

func (c *longConn) SetReadDeadline(t time.Time) error {
	err := c.Conn.SetReadDeadline(t)
	if err != nil {
		c.Lock()
		c.err = err
		c.Unlock()
	}
	return err
}

func (c *longConn) SetWriteDeadline(t time.Time) error {
	err := c.Conn.SetWriteDeadline(t)
	if err != nil {
		c.Lock()
		c.err = err
		c.Unlock()
	}
	return err
}

type ConnWithPkgSize struct {
	net.Conn
	Written int32
	Readn   int32
}

func (c *ConnWithPkgSize) Read(b []byte) (n int, err error) {
	n, err = c.Conn.Read(b)
	atomic.AddInt32(&(c.Readn), int32(n))
	return n, err
}

func (c *ConnWithPkgSize) Write(b []byte) (n int, err error) {
	n, err = c.Conn.Write(b)
	atomic.AddInt32(&(c.Written), int32(n))
	return n, err
}

func (c *ConnWithPkgSize) Close() error {
	err := c.Conn.Close()
	c.Conn = nil
	return err
}

// connRing is a struct managing all connections with same one address
type connRing struct {
	l    sync.Mutex
	arr  []*longConn
	size int
	tail int
	head int
}

func newConnRing(size int) *connRing {
	return &connRing{
		l:    sync.Mutex{},
		arr:  make([]*longConn, size+1),
		size: size,
		tail: 0,
		head: 0,
	}
}

// Push insert a PeerConn into ring
func (cr *connRing) Push(c *longConn) error {
	cr.l.Lock()
	if cr.isFull() {
		cr.l.Unlock()
		return errors.New("Ring is full")
	}
	cr.arr[cr.head] = c
	cr.head = cr.inc()
	cr.l.Unlock()
	return nil
}

// Pop return a PeerConn from ring, if ring is empty return nil
func (cr *connRing) Pop() *longConn {
	cr.l.Lock()
	if cr.isEmpty() {
		cr.l.Unlock()
		return nil
	}
	c := cr.arr[cr.tail]
	cr.arr[cr.tail] = nil
	cr.tail = cr.dec()
	cr.l.Unlock()
	return c
}

func (cr *connRing) inc() int {
	return (cr.head + 1) % (cr.size + 1)
}

func (cr *connRing) dec() int {
	return (cr.tail + 1) % (cr.size + 1)
}

func (cr *connRing) isEmpty() bool {
	return cr.tail == cr.head
}

func (cr *connRing) isFull() bool {
	return cr.inc() == cr.tail
}

// Peer has one address, it manage all connections base on this addresss
type peer struct {
	dialer Dialer
	lp     *LongPool

	addr           string
	ring           *connRing
	maxIdleConns   int
	maxIdleTimeout time.Duration
}

func newPeer(lp *LongPool, addr string, maxIdle int,
	maxIdleTimeout time.Duration,
	dialer Dialer) *peer {
	return &peer{
		lp:             lp,
		dialer:         dialer,
		addr:           addr,
		ring:           newConnRing(maxIdle),
		maxIdleConns:   maxIdle,
		maxIdleTimeout: maxIdleTimeout,
	}
}

// Get return a net.Conn from list
func (p *peer) Get(timeout time.Duration) (net.Conn, error) {
	// pick up connection from ring
	for {
		conn := p.ring.Pop()
		if conn == nil {
			break
		}
		p.lp.globalIdleDesc()
		if time.Now().Before(conn.deadline) {
			getAliveLongConnSucc(p.lp.targetPSM, p.addr)
			return conn, nil
		}
		// close connection after deadline
		conn.Conn.Close()
	}

	tcpConn, err := p.dialer.Dial(p.addr, timeout)
	if err != nil {
		newLongConnFail(p.lp.targetPSM, p.addr)
		return nil, err
	}
	newLongConnSucc(p.lp.targetPSM, p.addr)
	return &longConn{Conn: tcpConn, peer: p, deadline: time.Now().Add(p.maxIdleTimeout)}, nil
}

func (p *peer) put(c *longConn) error {
	if c.err != nil {
		longConnError(p.lp.targetPSM, p.addr)
		if c.Conn != nil {
			return c.Conn.Close()
		}
		return nil
	}
	c.deadline = time.Now().Add(p.maxIdleTimeout)
	if !p.lp.globalIdleInc() {
		return c.Conn.Close()
	}
	err := p.ring.Push(c)
	if err != nil {
		p.lp.globalIdleDesc()
		putLongConnFail(p.lp.targetPSM, p.addr)
		return c.Conn.Close()
	}
	putLongConnSucc(p.lp.targetPSM, p.addr)
	return nil
}

func (p *peer) Close() {
	for {
		conn := p.ring.Pop()
		if conn == nil {
			break
		}
		conn.Conn.Close()
	}
}

type LongPool struct {
	Dialer Dialer // hack dialer for test

	lock           sync.RWMutex
	peerMap        map[string]*peer
	maxIdlePerIns  int
	maxIdleGlobal  int
	maxIdleTimeout time.Duration
	targetPSM      string

	globalIdleCounter int64
}

// Get pick or generate a net.Conn and return
func (lp *LongPool) Get(host, port string, timeout time.Duration) (net.Conn, error) {
	addr := host + ":" + port
	p := lp.getPeer(addr)
	return p.Get(timeout)
}

func (lp *LongPool) getPeer(addr string) *peer {
	lp.lock.RLock()
	p, ok := lp.peerMap[addr]
	lp.lock.RUnlock()

	if ok {
		return p
	}
	p = newPeer(lp, addr, lp.maxIdlePerIns, lp.maxIdleTimeout, lp.Dialer)
	lp.lock.Lock()
	defer lp.lock.Unlock()
	if np, ok := lp.peerMap[addr]; ok { // created by other goroutine
		p = np
	} else {
		lp.peerMap[addr] = p
	}

	return p
}

func (lp *LongPool) Clean(host, port string) {
	addr := host + ":" + port
	lp.lock.Lock()
	p, ok := lp.peerMap[addr]
	delete(lp.peerMap, addr)
	lp.lock.Unlock()
	if ok {
		go p.Close()
	}
}

func (lp *LongPool) globalIdleInc() bool {
	if atomic.AddInt64(&lp.globalIdleCounter, 1) > int64(lp.maxIdleGlobal) {
		atomic.AddInt64(&lp.globalIdleCounter, -1)
		return false
	}
	return true
}

func (lp *LongPool) globalIdleDesc() {
	atomic.AddInt64(&lp.globalIdleCounter, -1)
}

// NewLongPool .
func NewLongPool(maxIdlePerIns, maxIdleGlobal int, maxIdleTimeout time.Duration, targetPSM string) *LongPool {
	return &LongPool{
		Dialer:         &innerDialer{},
		peerMap:        make(map[string]*peer, 30),
		maxIdleGlobal:  maxIdleGlobal,
		maxIdlePerIns:  maxIdlePerIns,
		maxIdleTimeout: maxIdleTimeout,
		targetPSM:      targetPSM,
	}
}
