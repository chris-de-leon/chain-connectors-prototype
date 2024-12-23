package flow_testutils

import "github.com/onflow/flow-go-sdk/access/grpc"

func InitBackend() (*grpc.Client, error) {
	return grpc.NewClient(grpc.EmulatorHost)
}
