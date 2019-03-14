package kiterrno

import (
	"fmt"

	"code.byted.org/gopkg/thrift"
)

const (
	// Thrift Application Exception
	UnknwonApllicationException = 0
	UnknownMethod               = 1
	InValidMessageTypeException = 2
	WrongMethodName             = 3
	BadSequenceID               = 4
	MissingResult               = 5
	InternalError               = 6
	ProtocolError               = 7

	UnknownErr       = 100
	ProxyInternalErr = 101

	ProxyConnTimeout  = 201
	ProxyReadTimeout  = 202
	ProxyWriteTimeout = 203
	CallConnTimeout   = 204
	CallReadTimeout   = 205
	CallWriteTimeout  = 206
	CallTotalTimeout  = 207

	// Protocal & Serialization
	BadProtocolErr = 301
	SerializeErr   = 302
	DeserializeErr = 303

	// Transport
	ProxyTransportErr  = 401
	CallerTransportErr = 402
	CalleeTransportErr = 403

	// Service Discovery
	ServiceDiscoveryInternalErr = 501
	ServiceDiscoveryEmptyErr    = 502

	// ACL
	NotAllowedByACL = 601

	// CB

	// Rate Limit
	OverConnectionLimit = 711
	OverQPSLimit        = 712
)

var defaultProxyExceptionMessage = map[int32]string{
	UnknwonApllicationException: "unknown application exception",
	UnknownMethod:               "unknown method",
	InValidMessageTypeException: "invalid message type",
	WrongMethodName:             "wrong method name",
	BadSequenceID:               "bad sequence ID",
	MissingResult:               "missing result",
	InternalError:               "unknown internal error",
	ProtocolError:               "unknown protocol error",

	UnknownErr:                  "unknown proxy error",
	ProxyInternalErr:            "proxy internal error",
	ProxyConnTimeout:            "proxy connection timeout",
	ProxyReadTimeout:            "proxy read timeout",
	ProxyWriteTimeout:           "proxy write timeout",
	CallConnTimeout:             "call connection timeout",
	CallReadTimeout:             "call read timeout",
	CallWriteTimeout:            "call write timeout",
	CallTotalTimeout:            "call total timeout",
	BadProtocolErr:              "bad protocal error",
	SerializeErr:                "serialize error",
	DeserializeErr:              "deserialize error",
	ProxyTransportErr:           "proxy transport error",
	CallerTransportErr:          "caller transport error",
	CalleeTransportErr:          "callee transport error",
	ServiceDiscoveryInternalErr: "service discovery internal error",
	ServiceDiscoveryEmptyErr:    "service discovery empty error",
	NotAllowedByACL:             "not allowed by acl",
	OverConnectionLimit:         "exceed endpoint connection limit",
	OverQPSLimit:                "exceed endpoint qps limit",
}

type ProxyException interface {
	thrift.TApplicationException
}

type proxyException struct {
	message string
	type_   int32
}

func (e proxyException) Error() string {
	prefix, exist := defaultProxyExceptionMessage[e.type_]
	if exist {
		if e.message != "" {
			return fmt.Sprintf("%s: %s", prefix, e.message)
		}
		return prefix
	}
	if e.message != "" {
		return e.message
	}
	return fmt.Sprintf("unknown proxy exception type: %d", e.type_)
}

func NewProxyException(type_ int32, message string) ProxyException {
	return &proxyException{message, type_}
}

func (p *proxyException) TypeId() int32 {
	return p.type_
}

func (p *proxyException) Read(iprot thrift.TProtocol) (thrift.TApplicationException, error) {
	_, err := iprot.ReadStructBegin()
	if err != nil {
		return nil, err
	}

	message := ""
	type_ := int32(thrift.UNKNOWN_APPLICATION_EXCEPTION)

	for {
		_, ttype, id, err := iprot.ReadFieldBegin()
		if err != nil {
			return nil, err
		}
		if ttype == thrift.STOP {
			break
		}
		switch id {
		case 1:
			if ttype == thrift.STRING {
				if message, err = iprot.ReadString(); err != nil {
					return nil, err
				}
			} else {
				if err = thrift.SkipDefaultDepth(iprot, ttype); err != nil {
					return nil, err
				}
			}
		case 2:
			if ttype == thrift.I32 {
				if type_, err = iprot.ReadI32(); err != nil {
					return nil, err
				}
			} else {
				if err = thrift.SkipDefaultDepth(iprot, ttype); err != nil {
					return nil, err
				}
			}
		default:
			if err = thrift.SkipDefaultDepth(iprot, ttype); err != nil {
				return nil, err
			}
		}
		if err = iprot.ReadFieldEnd(); err != nil {
			return nil, err
		}
	}
	return thrift.NewTApplicationException(type_, message), iprot.ReadStructEnd()
}

func (p *proxyException) Write(oprot thrift.TProtocol) (err error) {
	err = oprot.WriteStructBegin("TApplicationException")
	if len(p.Error()) > 0 {
		err = oprot.WriteFieldBegin("message", thrift.STRING, 1)
		if err != nil {
			return
		}
		err = oprot.WriteString(p.Error())
		if err != nil {
			return
		}
		err = oprot.WriteFieldEnd()
		if err != nil {
			return
		}
	}
	err = oprot.WriteFieldBegin("type", thrift.I32, 2)
	if err != nil {
		return
	}
	err = oprot.WriteI32(p.type_)
	if err != nil {
		return
	}
	err = oprot.WriteFieldEnd()
	if err != nil {
		return
	}
	err = oprot.WriteFieldStop()
	if err != nil {
		return
	}
	err = oprot.WriteStructEnd()
	return
}
