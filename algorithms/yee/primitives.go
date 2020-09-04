package yee

type RpcResponse struct {
	ID      int64       `json:"id"`
	JsonRpc string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
	Error   interface{} `json:"error"`
}

type Job struct {
	MerkleRoot       string
	ExtraData        [40]byte
	Target           string
	ShardCount       int16
	ShardBlockNumber map[string]uint64
}

type RpcError struct {
	code    int32
	message string
}
