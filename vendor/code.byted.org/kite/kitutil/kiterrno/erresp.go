package kiterrno

import "code.byted.org/kite/endpoint"

// GetErrCode extracts error code from response
func GetErrCode(resp interface{}) (int, bool) {
	kitResp, ok := resp.(endpoint.KitcCallResponse)
	if ok == false { // if resp is invalid, return directly
		return 0, false
	}

	baseResp := kitResp.GetBaseResp()
	if baseResp == nil {
		return 0, false
	}

	code := baseResp.GetStatusCode()
	return int(code), true
}

// error responses
var (
	// error resp for circuitbreaker MW
	ErrRespNotAllowedByServiceCB = NewErrResp(NotAllowedByServiceCBCode)
	ErrRespRPCTimeout            = NewErrResp(RPCTimeoutCode)

	// error resp for degradation MW
	ErrRespForbiddenByDegradation     = NewErrResp(ForbiddenByDegradationCode)
	ErrRespGetDegradationPercentError = NewErrResp(GetDegradationPercentErrorCode)

	// error resp for conn retry MW
	ErrRespConnRetry       = NewErrResp(ConnRetryCode)
	ErrRespBadConnBalancer = NewErrResp(BadConnBalancerCode)
	ErrRespBadConnRetrier  = NewErrResp(BadConnRetrierCode)

	// error resp for RPC retry MW
	ErrRespRPCRetry      = NewErrResp(RPCRetryCode)
	ErrRespBadPRCRetrier = NewErrResp(BadRPCRetrierCode)

	// error resp for common
	ErrRespNoExpectedField = NewErrResp(NoExpectedFieldCode)

	// error resp for pool
	ErrRespGetConnError = NewErrResp(GetConnErrorCode)

	// error resp for service discover
	ErrRespServiceDiscover = NewErrResp(ServiceDiscoverCode)

	// error resp for IDC selector
	ErrRespIDCSelectError = NewErrResp(IDCSelectErrorCode)

	// error resp for ACL
	ErrRespNotAllowedByACL = NewErrResp(NotAllowedByACLCode)

	// error resp for network timeout
	ErrRespReadTimeout     = NewErrResp(ReadTimeoutCode)
	ErrRespWriteTimeout    = NewErrResp(WriteTimeoutCode)
	ErrRespConnResetByPeer = NewErrResp(ConnRetryCode)
	ErrRespRemoteOrNet     = NewErrResp(RemoteOrNetErrCode)

	// error resp for QPS limit
	ErrRespEndpointQPSLimitReject = NewErrResp(EndpointQPSLimitRejectCode)

	// error resp for User Error circuitbreaker MW
	ErrRespNotAllowedByUserErrCB = NewErrResp(NotAllowedByUserErrCBCode)

	ErrStressBotRejection = NewErrResp(StressBotRejectionCode)
)

type baseResp struct {
	code int32
}

func (br *baseResp) GetStatusCode() int32 {
	return br.code
}

func (br *baseResp) GetStatusMessage() string {
	return ""
}

type callResp struct {
	baseResp   endpoint.KiteBaseResp
	remoteAddr string
}

func (cr *callResp) GetBaseResp() endpoint.KiteBaseResp {
	return cr.baseResp
}

func (cr *callResp) RemoteAddr() string {
	return cr.remoteAddr
}

func (cr *callResp) RealResponse() interface{} {
	return nil
}

// NewErrResp .
func NewErrResp(code int32) endpoint.KitcCallResponse {
	return &callResp{
		baseResp: &baseResp{
			code: code,
		},
	}
}

func NewErrRespWithAddr(code int32, addr string) endpoint.KitcCallResponse {
	return &callResp{
		baseResp: &baseResp{
			code: code,
		},
		remoteAddr: addr,
	}
}
