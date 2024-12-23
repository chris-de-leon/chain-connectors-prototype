package substrate_testutils

import gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"

const LOCAL_RPC_URL = "ws://host.docker.internal:9944"

func InitBackend() (*gsrpc.SubstrateAPI, error) {
	return gsrpc.NewSubstrateAPI(LOCAL_RPC_URL)
}
