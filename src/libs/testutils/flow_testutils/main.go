package flow_testutils

import "github.com/onflow/flow-go-sdk/access/grpc"

func InitBackend() (*grpc.Client, error) {
	client, err := grpc.NewClient(grpc.EmulatorHost)
	if err != nil {
		return nil, err
	} else {
		return client, nil
	}
}
