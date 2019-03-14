package kitc

import (
	"context"
	"errors"
	"io"
	"net"
	"time"

	"code.byted.org/gopkg/thrift"
)

const (
	bufferedTransportLen = 4096
)

type stringWriter interface {
	WriteString(s string) (int, error)
}

// TODO(xiangchao): Conn() net.Conn method should be added ?
type Transport interface {
	io.ReadWriteCloser
	io.ByteReader
	io.ByteWriter
	stringWriter
	Open() error
	IsOpen() bool
	Flush() error
	RemoteAddr() string
	OpenWithContext(ctx context.Context) error
}

func NewBufferedTransport(kc *KitcClient) Transport {
	if kc.meshMode {
		return NewMeshTHeaderBufferedTransportV1(kc)
	} else {
		return &BufferedTransport{client: kc}
	}
}

// BufferedTransport implement thrift.TRichTransport
type BufferedTransport struct {
	trans  *thrift.TBufferedTransport
	conn   net.Conn
	client *KitcClient
}

// RemoteAddr
func (bt *BufferedTransport) RemoteAddr() string {
	if bt.conn != nil {
		return bt.conn.RemoteAddr().String()
	}
	return ""
}

func (bt *BufferedTransport) Read(p []byte) (int, error) {
	return bt.trans.Read(p)
}

func (bt *BufferedTransport) Write(p []byte) (int, error) {
	return bt.trans.Write(p)
}

func (bt *BufferedTransport) ReadByte() (byte, error) {
	return bt.trans.ReadByte()
}

func (bt *BufferedTransport) WriteByte(c byte) error {
	return bt.trans.WriteByte(c)
}

func (bt *BufferedTransport) WriteString(s string) (int, error) {
	return bt.trans.WriteString(s)
}

func (bt *BufferedTransport) Flush() error {
	return bt.trans.Flush()
}

func (bt *BufferedTransport) OpenWithContext(ctx context.Context) error {
	rpcInfo := GetRPCInfo(ctx)
	conn := rpcInfo.Conn
	if conn == nil {
		return errors.New("No target connection in the context")
	}

	socket := GetSocketWithContext(conn, ctx)
	bt.trans = thrift.NewTBufferedTransport(socket, bufferedTransportLen)
	bt.conn = conn
	return nil
}

func (bt *BufferedTransport) Open() error {
	return bt.trans.Open()
}

func (bt *BufferedTransport) IsOpen() bool {
	return bt.trans.IsOpen()
}

func (bt *BufferedTransport) Close() error {
	return bt.trans.Close()
}

// NewFramedTransport return a FramedTransport
func NewFramedTransport(kc *KitcClient) Transport {
	if kc.meshMode {
		return NewMeshTHeaderFramedTransportV1(kc)
	} else {
		return &FramedTransport{
			client:    kc,
			maxLength: kc.maxFramedSize,
		}
	}
}

// FramedTransport implement thrift.TRichTransport
type FramedTransport struct {
	trans     *thrift.TFramedTransport
	conn      net.Conn
	client    *KitcClient
	maxLength int32
}

func (ft *FramedTransport) RemoteAddr() string {
	if ft.conn != nil {
		return ft.conn.RemoteAddr().String()
	}
	return ""
}

func (ft *FramedTransport) Read(p []byte) (int, error) {
	return ft.trans.Read(p)
}

func (ft *FramedTransport) Write(p []byte) (int, error) {
	return ft.trans.Write(p)
}

func (ft *FramedTransport) ReadByte() (byte, error) {
	return ft.trans.ReadByte()
}

func (ft *FramedTransport) WriteByte(c byte) error {
	return ft.trans.WriteByte(c)
}

func (ft *FramedTransport) WriteString(s string) (int, error) {
	return ft.trans.WriteString(s)
}

func (ft *FramedTransport) Flush() error {
	return ft.trans.Flush()
}

func (ft *FramedTransport) Open() error {
	return ft.trans.Open()
}

// OpenWithContext connect a backend server acording the content of ctx
func (ft *FramedTransport) OpenWithContext(ctx context.Context) error {
	rpcInfo := GetRPCInfo(ctx)
	conn := rpcInfo.Conn
	if conn == nil {
		return errors.New("No target connection in the context")
	}

	socket := GetSocketWithContext(conn, ctx)
	bt := thrift.NewTBufferedTransport(socket, bufferedTransportLen)
	ft.trans = thrift.NewTFramedTransportMaxLength(bt, int(ft.maxLength))
	ft.conn = conn
	return nil
}

func (ft *FramedTransport) IsOpen() bool {
	return ft.trans.IsOpen()
}

func (ft *FramedTransport) Close() error {
	return ft.trans.Close()
}

// func GetConnWithContext(client *KitcClient, ctx context.Context) (net.Conn, error) {
// 	return client.Pool.Get(ctx)
// }

// GetSocketWithContext .
func GetSocketWithContext(conn net.Conn, ctx context.Context) *thrift.TSocket {
	// TODO(zhangyuanjia): use the max one between readtimeout and writetimeout ?
	rpcInfo := GetRPCInfo(ctx)
	timeout := time.Duration(rpcInfo.WriteTimeout) * time.Millisecond
	if rpcInfo.ReadTimeout > rpcInfo.WriteTimeout {
		timeout = time.Duration(rpcInfo.ReadTimeout) * time.Millisecond
	}

	deadline, ok := ctx.Deadline()
	if dur := deadline.Sub(time.Now()); ok && dur < timeout {
		timeout = dur
	}

	return thrift.NewTSocketFromConnTimeout(conn, timeout)
}

// GetSocketWithContext .
func MeshGetSocketWithContext(conn net.Conn, ctx context.Context) *thrift.TSocket {
	timeout := time.Duration(defaultMeshProxyConfig.WriteTimeout) * time.Millisecond
	if defaultMeshProxyConfig.ReadTimeout > defaultMeshProxyConfig.WriteTimeout {
		timeout = time.Duration(defaultMeshProxyConfig.ReadTimeout) * time.Millisecond
	}

	deadline, ok := ctx.Deadline()
	if dur := deadline.Sub(time.Now()); ok && dur < timeout {
		timeout = dur
	}

	return thrift.NewTSocketFromConnTimeout(conn, timeout)
}

// NOTICE: 为了兼容，client Request是HeaderTransport，Response是原来Transport(Buffered/Framed)
// FIXME: thrift.TBufferedTransport和thrift.TFrameredTransport没有合适的公共接口，所以实现了两份
type MeshTHeaderBufferedTransportV1 struct {
	reqTrans  *thrift.HeaderTransport
	resTrans  *thrift.TBufferedTransport
	conn   net.Conn
	client *KitcClient
}

func NewMeshTHeaderBufferedTransportV1(kc *KitcClient) *MeshTHeaderBufferedTransportV1 {
	return &MeshTHeaderBufferedTransportV1{
		client: kc,
	}
}

// RemoteAddr
func (bt *MeshTHeaderBufferedTransportV1) RemoteAddr() string {
	if bt.conn != nil {
		return bt.conn.RemoteAddr().String()
	}
	return ""
}

func (bt *MeshTHeaderBufferedTransportV1) Read(p []byte) (int, error) {
	return bt.resTrans.Read(p)
}

func (bt *MeshTHeaderBufferedTransportV1) ReadByte() (byte, error) {
	return bt.resTrans.ReadByte()
}

func (bt *MeshTHeaderBufferedTransportV1) Write(p []byte) (int, error) {
	return bt.reqTrans.Write(p)
}

func (bt *MeshTHeaderBufferedTransportV1) WriteByte(c byte) error {
	return bt.reqTrans.WriteByte(c)
}

func (bt *MeshTHeaderBufferedTransportV1) WriteString(s string) (int, error) {
	return bt.reqTrans.WriteString(s)
}

func (bt *MeshTHeaderBufferedTransportV1) Flush() error {
	return bt.reqTrans.Flush()
}

func (bt *MeshTHeaderBufferedTransportV1) OpenWithContext(ctx context.Context) error {
	rpcInfo := GetRPCInfo(ctx)
	conn := rpcInfo.Conn
	if conn == nil {
		return errors.New("No target connection in the context")
	}

	socket := MeshGetSocketWithContext(conn, ctx)
	bt.reqTrans = thrift.NewHeaderTransport(socket)
	bt.reqTrans.SetClientType(thrift.HeaderUnframedClientType)
	if bt.client.opts.ProtocolType == ProtocolBinary {
		bt.reqTrans.SetProtocolID(thrift.ProtocolIDBinary)
	} else {
		bt.reqTrans.SetProtocolID(thrift.ProtocolIDCompact)
	}
	bt.resTrans = thrift.NewTBufferedTransport(socket, bufferedTransportLen)
	bt.conn = conn
	bt.reqTrans.SetIntHeader(MESH_VERSION, MeshTHeaderProtocolVersion)
	bt.reqTrans.SetIntHeader(TRANSPORT_TYPE, "unframed")
	if intHeaders, ok := ctx.Value(THeaderInfoIntHeaders).(map[uint16]string); ok {
		for k, v := range intHeaders {
			bt.reqTrans.SetIntHeader(k, v)
		}
	}
	if headers, ok := ctx.Value(THeaderInfoHeaders).(map[string]string); ok {
		for k, v := range headers {
			bt.reqTrans.SetHeader(k, v)
		}
	}
	return nil
}

func (bt *MeshTHeaderBufferedTransportV1) Open() error {
	return bt.reqTrans.Open()
}

func (bt *MeshTHeaderBufferedTransportV1) IsOpen() bool {
	return bt.reqTrans.IsOpen()
}

func (bt *MeshTHeaderBufferedTransportV1) Close() error {
	return bt.resTrans.Close()
}

type MeshTHeaderFramedTransportV1 struct {
	reqTrans  *thrift.HeaderTransport
	resTrans  *thrift.TFramedTransport
	conn      net.Conn
	client    *KitcClient
	maxLength int32
}

func NewMeshTHeaderFramedTransportV1(kc *KitcClient) *MeshTHeaderFramedTransportV1 {
	return &MeshTHeaderFramedTransportV1{
		client:    kc,
		maxLength: kc.maxFramedSize,
	}
}

// RemoteAddr
func (ft *MeshTHeaderFramedTransportV1) RemoteAddr() string {
	if ft.conn != nil {
		return ft.conn.RemoteAddr().String()
	}
	return ""
}

func (ft *MeshTHeaderFramedTransportV1) Read(p []byte) (int, error) {
	return ft.resTrans.Read(p)
}

func (ft *MeshTHeaderFramedTransportV1) ReadByte() (byte, error) {
	return ft.resTrans.ReadByte()
}

func (ft *MeshTHeaderFramedTransportV1) Write(p []byte) (int, error) {
	return ft.reqTrans.Write(p)
}

func (ft *MeshTHeaderFramedTransportV1) WriteByte(c byte) error {
	return ft.reqTrans.WriteByte(c)
}

func (ft *MeshTHeaderFramedTransportV1) WriteString(s string) (int, error) {
	return ft.reqTrans.WriteString(s)
}

func (ft *MeshTHeaderFramedTransportV1) Flush() error {
	return ft.reqTrans.Flush()
}

func (ft *MeshTHeaderFramedTransportV1) OpenWithContext(ctx context.Context) error {
	rpcInfo := GetRPCInfo(ctx)
	conn := rpcInfo.Conn
	if conn == nil {
		return errors.New("No target connection in the context")
	}

	socket := MeshGetSocketWithContext(conn, ctx)
	ft.reqTrans = thrift.NewHeaderTransport(socket)
	ft.reqTrans.SetClientType(thrift.HeaderFramedClientType)
	if ft.client.opts.ProtocolType == ProtocolBinary {
		ft.reqTrans.SetProtocolID(thrift.ProtocolIDBinary)
	} else {
		ft.reqTrans.SetProtocolID(thrift.ProtocolIDCompact)
	}
	bt := thrift.NewTBufferedTransport(socket, bufferedTransportLen)
	ft.resTrans = thrift.NewTFramedTransportMaxLength(bt, int(ft.maxLength))
	ft.conn = conn
	ft.reqTrans.SetIntHeader(MESH_VERSION, MeshTHeaderProtocolVersion)
	ft.reqTrans.SetIntHeader(TRANSPORT_TYPE, "framed")
	if intHeaders, ok := ctx.Value(THeaderInfoIntHeaders).(map[uint16]string); ok {
		for k, v := range intHeaders {
			ft.reqTrans.SetIntHeader(k, v)
		}
	}
	if headers, ok := ctx.Value(THeaderInfoHeaders).(map[string]string); ok {
		for k, v := range headers {
			ft.reqTrans.SetHeader(k, v)
		}
	}
	return nil
}

func (ft *MeshTHeaderFramedTransportV1) Open() error {
	return ft.reqTrans.Open()
}

func (ft *MeshTHeaderFramedTransportV1) IsOpen() bool {
	return ft.reqTrans.IsOpen()
}

func (ft *MeshTHeaderFramedTransportV1) Close() error {
	return ft.resTrans.Close()
}