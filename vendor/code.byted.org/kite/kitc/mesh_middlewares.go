/*
@zhengjianbo: mesh相关中间件，非mesh功能不要添加
*/
package kitc

import (
	"fmt"
	"strconv"
	"time"
	"context"

	"code.byted.org/gopkg/env"
	"code.byted.org/gopkg/thrift"
	"code.byted.org/kite/endpoint"
	"code.byted.org/kite/kitc/connpool"
	"code.byted.org/kite/kitutil/kiterrno"
)

// TODO(zhanggongyuan): remember to add opentracing instrument to record conn-event
// TODO(zhengjianbo): mesh场景下只直连proxy，pool可定制化下
func NewMeshPoolMW(pooler connpool.ConnPool) endpoint.Middleware {
	return func(next endpoint.EndPoint) endpoint.EndPoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			rpcInfo := GetRPCInfo(ctx)
			begin := time.Now()

			conn, err := pooler.Get(rpcInfo.TargetInstance.Host,
				rpcInfo.TargetInstance.Port,
				time.Duration(defaultMeshProxyConfig.ConnectTimeout)*time.Millisecond)

			rpcInfo.ConnCost = time.Now().Sub(begin) / time.Microsecond
			if err != nil {
				kerr := kiterrno.NewKitErr(kiterrno.GetConnErrorCode, err)
				return kiterrno.ErrRespGetConnError, kerr
			}

			rpcInfo.Conn = &connpool.ConnWithPkgSize{
				Conn: conn,
			}
			return next(ctx, request)
		}
	}
}

// Mesh THeader set Headers .
func MeshSetHeadersMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		r := GetRPCInfo(ctx)
		headers := map[uint16]string{}
		headers[LOG_ID] = r.LogID // logid
		headers[ENV] = r.Env
		headers[FROM_SERVICE] = r.From        // from_service
		headers[FROM_CLUSTER] = r.FromCluster // from_cluster
		headers[FROM_IDC] = env.IDC()     // from_idc
		headers[TO_SERVICE] = r.To          // to_service
		headers[TO_METHOD] = r.Method      // to_method
		// optional field
		if len(r.ToCluster) > 0 {
			headers[TO_CLUSTER] = r.ToCluster // to_cluster
		}
		if len(r.TargetIDC) > 0 {
			headers[TO_IDC] = r.TargetIDC // to_idc
		}
		if len(r.Instances) > 0 {
			// only support one
			// dest_address
			headers[DEST_ADDRESS] = r.Instances[0].Host + ":" + r.Instances[0].Port
		}
		// 用户指定超时配置rt, ct, rdt, wrt
		if r.RPCTimeout >= 0 {
			headers[RPC_TIMEOUT] = strconv.Itoa(r.RPCTimeout) // rpc_timeout
		}
		if r.ReadTimeout >= 0 {
			headers[READ_TIMEOUT] = strconv.Itoa(r.ReadTimeout) // read_timeout
		}
		// mesh场景下conn_timeout由ms配置，write_timeout没有意义
		if len(r.RingHashKey) > 0 {
			headers[RING_HASH_KEY] = r.RingHashKey
		}
		if len(r.DDPTag) > 0 {
			headers[DDP_TAG] = r.DDPTag
		}
		ctx = context.WithValue(ctx, THeaderInfoIntHeaders, headers)

		return next(ctx, request)
	}
}

// TODO(zhanggongyuan): remember to add opentracing instrument to record send-recv pkg event
// MeshIOErrorHandlerMW .
func MeshIOErrorHandlerMW(next endpoint.EndPoint) endpoint.EndPoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		resp, err := next(ctx, request)
		if err == nil {
			return resp, nil
		}
		if terr, ok := err.(thrift.TApplicationException); ok {
			perr := kiterrno.NewProxyException(terr.TypeId(), terr.Error())
			return kiterrno.NewErrResp(terr.TypeId()), perr
		}
		return kiterrno.ErrRespRemoteOrNet, fmt.Errorf("remote or network err: %s", err.Error())
	}
}
