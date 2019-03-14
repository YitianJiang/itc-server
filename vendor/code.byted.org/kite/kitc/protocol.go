package kitc

type TransCtxKey string
const (
	THeaderInfoIntHeaders = TransCtxKey("theader-int-headers")
	THeaderInfoHeaders    = TransCtxKey("theader-headers")
)

type ProtocolType int
const (
	ProtocolBinary   ProtocolType = 0
	ProtocolCompact  ProtocolType = 1
)

// mesh使用的协议版本号
const MeshTHeaderProtocolVersion = "1.0.0"

const (
	MESH_VERSION uint16 = iota
	TRANSPORT_TYPE
	LOG_ID
	FROM_SERVICE
	FROM_CLUSTER
	FROM_IDC
	TO_SERVICE
	TO_CLUSTER
	TO_IDC
	TO_METHOD
	ENV
	DEST_ADDRESS
	RPC_TIMEOUT
	READ_TIMEOUT
	RING_HASH_KEY
	DDP_TAG
)
