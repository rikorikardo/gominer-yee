package yee

import "testing"

func Test_decodeWork(t *testing.T) {
	buf := []byte(`{"jsonrpc":"2.0","result":{"extra_data":[0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,159,14,68,76],"merkle_root":"0xe37ffba78fd0a3dd86ee72e9f5b5c56f14b2a9a0fbde63eb1e394c8cbd077aa1","shard_block_number":{"0":68943,"1":69057,"2":69055,"3":69055},"shard_count":4,"target":"0x271cf8d64287e4f549ad65c2ce05a716f3eb17190c913cbee5736d"},"id":1}`)

	var cl = &RpcClient{"http://127.0.0.1:10033"}
	target, header, err := cl.decodeWork(buf)
	if err != nil {
		t.Error(err.Error())
	}
	println(target)
	println(header)

}
