package kiterrno

// KitErr .
type KitErr struct {
	underlying error
	errno      int
}

// NewKitErr .
func NewKitErr(errno int, err error) error {
	return &KitErr{
		underlying: err,
		errno:      errno,
	}
}

func (ke *KitErr) Error() string {
	prefix := ""
	switch ke.errno {
	case NotAllowedByServiceCBCode:
		prefix = NotAllowedByServiceCBDesc
	case NotAllowedByInstanceCBCode:
		prefix = NotAllowedByInstanceCBDesc
	case ForbiddenByDegradationCode:
		prefix = ForbiddenByDegradationDesc
	case GetDegradationPercentErrorCode:
		prefix = GetDegradationPercentErrorDesc
	case BadConnBalancerCode:
		prefix = BadConnBalancerDesc
	case BadConnRetrierCode:
		prefix = BadConnRetrierDesc
	case ConnRetryCode:
		prefix = ConnRetryDesc
	case NoExpectedFieldCode:
		prefix = NoExpectedFieldDesc
	case GetConnErrorCode:
		prefix = GetConnErrorDesc
	case RPCTimeoutCode:
		prefix = RPCTimeoutDesc
	case ServiceDiscoverCode:
		prefix = ServiceDiscoverDesc
	case IDCSelectErrorCode:
		prefix = IDCSelectErrorDesc
	case RPCRetryCode:
		prefix = RPCRetryDesc
	case BadRPCRetrierCode:
		prefix = BadRPCRetrierDesc
	case NotAllowedByACLCode:
		prefix = NotAllowedByACLDesc
	case ReadTimeoutCode:
		prefix = ReadTimeoutDesc
	case WriteTimeoutCode:
		prefix = WriteTimeoutDesc
	case ConnResetByPeerCode:
		prefix = ConnResetByPeerDesc
	case StressBotRejectionCode:
		prefix = StressBotRejectionDesc
	case EndpointQPSLimitRejectCode:
		prefix = EndpointQPSLimitRejectDesc
	case NotAllowedByUserErrCBCode:
		prefix = NotAllowedByUserErrCBDesc
	}
	if ke.underlying != nil {
		return prefix + ": " + ke.underlying.Error()
	}
	return prefix
}

// Underlying .
func (ke *KitErr) Underlying() error {
	return ke.underlying
}

// Errno .
func (ke *KitErr) Errno() int {
	return ke.errno
}
