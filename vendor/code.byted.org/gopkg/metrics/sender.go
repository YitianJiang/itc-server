package metrics

import (
	"io"
	"net"
	"strings"
	"sync"
	"time"
	"unsafe"
)

const (
	// DO NOT MODIFY IT IF YOU DONT KNOWN WHAT YOU ARE DOING
	maxBunchBytes = 32 << 10 // 32kb

	// send metrics immediately if larger than the size
	maxPendingSize = 1000

	// 200ms timeout before send metrics
	emitInterval = 200 * time.Millisecond
)

const (
	_emit = "emit"
)

var unixdomainsock = ""

func init() {
	for _, path := range []string{"/opt/tmp/sock/metric.sock", "/tmp/metric.sock"} {
		for i := 0; i < 3; i++ {
			conn, err := net.Dial("unixgram", path)
			if err == nil {
				conn.Close()
				unixdomainsock = path
				return
			}
			if strings.Contains(err.Error(), "no such file") {
				break
			}
		}
	}
}

type metricEntry struct {
	mt     metricsType
	prefix string
	name   string
	v      float64
	ts     int64

	tt *cachedTags
}

func (m *metricEntry) MarshalSize() int {
	// protocol: 6 fields: emit $type $prefix.name  $value $tag ""
	n := 0
	n += msgpackArrayHeaderSize
	n += msgpackStringSize(_emit)
	n += msgpackStringSize(m.mt.String())
	if len(m.prefix) > 0 {
		n += msgpackStringHeaderSize + (len(m.prefix) + 1 + len(m.name))
	} else {
		n += msgpackStringSize(m.name)
	}
	n += msgpackStringHeaderSize + floatStrSize(m.v) // int64 + "." + 5 prec float + str header
	n += msgpackStringHeaderSize + len(m.tt.Bytes())
	if m.ts > 0 {
		n += msgpackStringHeaderSize + int64StrSize(m.ts)
	} else {
		n += msgpackStringHeaderSize + 0
	}
	return n
}

func (m *metricEntry) AppendTo(p []byte) []byte {
	// protocol: 6 fields: emit $type $prefix.name  $value $tag ""
	p = msgpackAppendArrayHeader(p, 6)
	p = msgpackAppendString(p, _emit)
	p = msgpackAppendString(p, m.mt.String())
	if len(m.prefix) > 0 {
		p = msgpackAppendStringHeader(p, uint16(len(m.prefix)+1+len(m.name)))
		p = append(p, m.prefix...)
		p = append(p, '.')
		p = append(p, m.name...)
	} else {
		p = msgpackAppendString(p, m.name)
	}
	p = msgpackAppendStringHeader(p, uint16(floatStrSize(m.v)))
	p = appendFloat64(p, m.v)
	p = msgpackAppendStringHeader(p, uint16(len(m.tt.Bytes())))
	p = append(p, m.tt.Bytes()...)
	if m.ts > 0 {
		p = msgpackAppendStringHeader(p, uint16(int64StrSize(m.ts)))
		p = appendInt64(p, m.ts)
	} else {
		p = msgpackAppendString(p, "")
	}
	return p
}

func (m *metricEntry) MarshalTo(b []byte) {
	p := b[:0]
	p = m.AppendTo(p)
	if len(p) != len(b) {
		panic("buf size err")
	}
}

type metricsWriter struct {
	mu    sync.RWMutex
	addr  string
	err   error
	conn  net.Conn
	ctime time.Time
}

func (w *metricsWriter) Write(b []byte) (n int, err error) {
	if w.addr == BlackholeAddr {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	n = len(b)
	if w.conn == nil {
		now := time.Now()
		if now.Sub(w.ctime) < time.Second {
			return
		}
		w.ctime = now
		if strings.HasPrefix(w.addr, "/") {
			w.conn, err = net.Dial("unixgram", w.addr)
		} else {
			w.conn, err = net.Dial("udp", w.addr)
		}
		if err != nil {
			println("metrics conn err: ", err.Error())
			w.conn = nil
			return
		}
	}
	_, err = w.conn.Write(b)
	if err != nil {
		w.conn.Close()
		w.conn = nil
	}
	return
}

type Sender struct {
	addr      string
	makeBunch bool

	w io.Writer

	counterChan chan metricEntry
	otherChan   chan metricEntry
}

func NewSender(addr string) *Sender {
	s := &Sender{addr: addr}
	if addr == DefaultMetricsServer && unixdomainsock != "" {
		s.addr = unixdomainsock
		s.makeBunch = true
	}
	s.w = &metricsWriter{addr: s.addr}
	s.counterChan = make(chan metricEntry, 4*maxPendingSize)
	s.otherChan = make(chan metricEntry, 4*maxPendingSize)
	s.runLoops()
	return s
}

func (s *Sender) SendAsync(m metricEntry) error {
	ch := s.otherChan
	if m.mt == metricsTypeCounter || m.mt == metricsTypeRateCounter {
		ch = s.counterChan
	}
	select {
	case ch <- m:
	default:
		return ErrEmitBufferFull
	}
	return nil
}

func (s *Sender) SendAsyncBlock(m metricEntry) error {
	ch := s.otherChan
	if m.mt == metricsTypeCounter || m.mt == metricsTypeRateCounter {
		ch = s.counterChan
	}
	ch <- m
	return nil
}

func (s *Sender) runLoops() {
	go s.emitCounterLoop()
	go s.emitOtherLoop()
}

func (s *Sender) emitCounterLoop() {
	tick := time.Tick(emitInterval)
	p := make([]metricEntry, 0, maxPendingSize)
	for {
		select {
		case e := <-s.counterChan:
			p = append(p, e)
			if len(p) >= maxPendingSize {
				s.SendCounter(p)
				p = p[:0]
			}
		case <-tick:
			if len(p) > 0 {
				s.SendCounter(p)
				p = p[:0]
			}
		}
	}
}

func (s *Sender) emitOtherLoop() {
	tick := time.Tick(emitInterval)
	p := make([]metricEntry, 0, maxPendingSize)
	for {
		select {
		case e := <-s.otherChan:
			p = append(p, e)
			if len(p) >= maxPendingSize {
				s.Send(p)
				p = p[:0]
			}
		case <-tick:
			if len(p) > 0 {
				s.Send(p)
				p = p[:0]
			}
		}
	}
}

type aggregatekey struct {
	prefix string
	name   string
	tt     uintptr // tags pointer
	mt     metricsType
}

type counterAggregator struct {
	keys []aggregatekey
	m    map[aggregatekey]*metricEntry
}

var counterAggregatorPool = sync.Pool{
	New: func() interface{} {
		return &counterAggregator{
			keys: make([]aggregatekey, 0, maxPendingSize),
			m:    make(map[aggregatekey]*metricEntry, maxPendingSize),
		}
	},
}

func (a *counterAggregator) Merge(ms []metricEntry) []metricEntry {
	for i := range ms {
		e := &ms[i]
		k := aggregatekey{
			prefix: e.prefix,
			name:   e.name,
			tt:     uintptr(unsafe.Pointer(e.tt)),
			mt:     e.mt,
		}
		v, ok := a.m[k]
		if ok {
			v.v += e.v
		} else {
			a.m[k] = e
			a.keys = append(a.keys, k)
		}
	}
	p := ms[:0]
	for _, k := range a.keys {
		p = append(p, *a.m[k])
		delete(a.m, k)
	}
	a.keys = a.keys[:0]
	return p
}

func (s *Sender) SendCounter(ms []metricEntry) {
	if len(ms) == 0 {
		return
	}
	a := counterAggregatorPool.Get().(*counterAggregator)
	s.Send(a.Merge(ms))
	counterAggregatorPool.Put(a)
}

func (s *Sender) Send(ms []metricEntry) {
	if !s.makeBunch {
		p := wbufpool.Get().(*wbuf)
		defer wbufpool.Put(p)
		for _, m := range ms {
			s.w.Write(m.AppendTo(p.b[:0]))
		}
		return
	}
	// send bunch
	for len(ms) > 0 {
		ms = ms[s.sendbunch(ms):]
	}
}

type wbuf struct {
	b []byte

	mem [2 * maxBunchBytes]byte
}

var wbufpool = sync.Pool{
	New: func() interface{} {
		p := new(wbuf)
		p.b = p.mem[:0]
		return p
	},
}

func (s *Sender) sendbunch(ms []metricEntry) int {
	if len(ms) == 0 {
		return 0
	}

	// limit to send maxBunchBytes
	k := 0
	n := msgpackArrayHeaderSize
	for _, m := range ms {
		n += m.MarshalSize()
		if n >= maxBunchBytes {
			break
		}
		k++
	}

	if k == 0 {
		panic("metrics too large to send") // single metrics > maxBunchBytes
	}

	ms = ms[:k]

	p := wbufpool.Get().(*wbuf)
	defer wbufpool.Put(p)
	p.b = p.b[:0]

	// marshal to p
	p.b = msgpackAppendArrayHeader(p.b, uint16(len(ms)))
	for _, m := range ms {
		p.b = m.AppendTo(p.b)
	}
	s.w.Write(p.b)
	return k
}
