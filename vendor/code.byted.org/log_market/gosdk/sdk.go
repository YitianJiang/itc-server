package gosdk

import (
	"net"
	"time"
	"fmt"
	"os"
	"github.com/juju/ratelimit"
	"code.byted.org/gopkg/metrics"
	"github.com/gogo/protobuf/proto"
	"strconv"
	"sync/atomic"
	"errors"
	"sync"
)

var (
	debugMode          bool
	raddr              *net.UnixAddr
	metricsClient      *metrics.MetricsClient
	errCount           = int64(0)
	ch                 = make(chan *Msg, 1024*4)
	network            = "unixpacket"
	socketPath         = "/opt/tmp/ttlogagent/unixpacket_v2.sock"
	sendSuccessTag     = map[string]string{"status": "success"}
	connRateLimitTag   = map[string]string{"status": "error", "reason": "connRateLimit"}
	connWriteErrorTag  = map[string]string{"status": "error", "reason": "connWriteError"}
	asyncWriteErrorTag = map[string]string{"status": "error", "reason": "asyncWrite"}
	connErrorTag       = map[string]string{"status": "error", "reason": "connError"}
	marshalErrorTag    = map[string]string{"status": "error", "reason": "marshalError"}
	ErrChannelFull     = errors.New("[logagent-gosdk] channel full")
	ErrMsgNil          = errors.New("msg cannot be nil")
	ErrStop            = errors.New("gosdk had exited gracefully ")
	status             = int32(0)
	exitCh             = make(chan struct{})
	wg                 = sync.WaitGroup{}
)

const (
	statusStop    = iota
	statusRunning
)

func init() {
	var err error
	raddr, err = net.ResolveUnixAddr("unixpacket", socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[logagent-gosdk] init error:%v", err)
		return
	}
	mode := os.Getenv("__sdk_mode")
	if mode == "debug" {
		fmt.Println("[logagent-gosdk] use debug mode")
		debugMode = true
	}
	metricsClient = metrics.NewDefaultMetricsClient("toutiao.infra.ttlogagent.gosdk", true)
	status = statusRunning
	workerNum := 3
	wg.Add(workerNum)
	for i := 0; i < workerNum; i ++ {
		go NewSender().run()
	}
}

func Send(taskName string, m *Msg) error {
	if atomic.LoadInt32(&status) != statusRunning {
		return ErrStop
	}
	if m == nil {
		return ErrMsgNil
	}
	tags := make(map[string]string)
	for k, v := range m.Tags {
		tags[k] = v
	}
	tags["_taskName"] = taskName
	msgCopy := &Msg{
		Msg:  m.Msg, //do not copy for performance
		Tags: tags,
	}
	select {
	case ch <- msgCopy:
		return nil
	default:
		atomic.AddInt64(&errCount, 1)
		printToStderrIfDebug("[logagent-gosdk] discard message " + strconv.Itoa(int(atomic.LoadInt64(&errCount))))
		metricsClient.EmitCounter(taskName+".send", 1, "", asyncWriteErrorTag)
		return ErrChannelFull
	}
}

func GracefullyExit() {
	if !atomic.CompareAndSwapInt32(&status, statusRunning, statusStop) {
		return
	}
	close(exitCh)
	wg.Wait()
}

type sender struct {
	conn            net.Conn
	connRetryLimit  *ratelimit.Bucket
	packetSizeLimit int
	batchSizeByte   int
	batchTimeoutMs  int
	lastSendTime    time.Time
	batch           *MsgBatch
}

func NewSender() *sender {
	return &sender{
		connRetryLimit: ratelimit.NewBucket(3*time.Second, 1),
		//unix socket buffer size is :  160K on 32-bit linux, 208K on 64-bit linux. packetSizeLimit must be less than above.
		//and 7~8K packet will be more friendly for memory allocate.
		packetSizeLimit: 7 * 1024,
		batchSizeByte:   0,
		batchTimeoutMs:  200,
		lastSendTime:    time.Now(),
		batch: &MsgBatch{
			Msgs: make([]*Msg, 0, 256),
		},
	}
}

func (s *sender) run() {
	if c, err := net.DialUnix(network, nil, raddr); err == nil {
		s.conn = c
	}
	ticker := time.Tick(time.Millisecond * time.Duration(s.batchTimeoutMs))
	for {
		select {
		case msg := <-ch:
			s.batch.Msgs = append(s.batch.Msgs, msg)
			s.batchSizeByte += msg.Size()
			if s.batchSizeByte >= s.packetSizeLimit {
				s.flush()
			}
		case <-ticker:
			if int(time.Now().Sub(s.lastSendTime))/int(time.Millisecond) > s.batchTimeoutMs {
				s.flush()
			}
		case <-exitCh:
			chanLen := len(ch)
			for i := 0; i < chanLen; i++ {
				select {
				case msg1 := <-ch:
					s.batch.Msgs = append(s.batch.Msgs, msg1)
				default:
				}
			}
			s.flush()
			wg.Done()
			return
		}
	}
}

func (s *sender) send(buf []byte) (error, map[string]string) {
	if s.conn == nil {
		if s.connRetryLimit.TakeAvailable(1) < 1 {
			return fmt.Errorf("[logagent-gosdk] build connection error=rate limit"), connRateLimitTag
		}
		c, err := net.DialUnix(network, nil, raddr)
		if err != nil {
			return fmt.Errorf("[logagent-gosdk] build connection error=%v", err), connErrorTag
		}
		s.conn = c
	}
	s.conn.SetWriteDeadline(time.Now().Add(time.Millisecond * 100))
	if _, err := s.conn.Write(buf); err != nil {
		s.conn.Close()
		s.conn = nil
		return fmt.Errorf("[logagent-gosdk] send batch error=%v", err), connWriteErrorTag
	}
	return nil, nil
}

func (s *sender) flush() {
	if len(s.batch.Msgs) == 0 {
		return
	}
	defer func() {
		s.batchSizeByte = 0
		s.batch.Msgs = s.batch.Msgs[:0]
		s.lastSendTime = time.Now()
	}()

	buf, err := proto.Marshal(s.batch)
	if err != nil {
		printToStderrIfDebug(fmt.Sprintf("[logagent-gosdk] proto marashal batch error=%v", err))
		atomic.AddInt64(&errCount, int64(len(s.batch.Msgs)))
		for _, msg := range s.batch.Msgs {
			metricsClient.EmitCounter(msg.Tags["_taskName"]+".send", 1, "", marshalErrorTag)
		}
		return
	}
	if err, errorTag := s.send(buf); err != nil {
		printToStderrIfDebug("first send failed :" + err.Error())
		if err, errorTag = s.send(buf); err != nil {
			atomic.AddInt64(&errCount, int64(len(s.batch.Msgs)))
			for _, msg := range s.batch.Msgs {
				metricsClient.EmitCounter(msg.Tags["_taskName"]+".send", 1, "", errorTag)
			}
			printToStderrIfDebug(err.Error())
			return
		}
	}
	for _, msg := range s.batch.Msgs {
		metricsClient.EmitCounter(msg.Tags["_taskName"]+".send", 1, "", sendSuccessTag)
	}
}

func printToStderrIfDebug(msg string) {
	if !debugMode {
		return
	}
	fmt.Fprintln(os.Stderr, fmt.Sprintf("error count:%v ,%v ", atomic.LoadInt64(&errCount), msg))
}
