package solana_testutils

import (
	"context"
	"time"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

var (
	DefaultConfirmationTimeout = time.Second * 30
	DefaultInitFundAmountInSol = 100
)

type Backend struct {
	RpcClient *rpc.Client
	WssClient *ws.Client
}

func InitBackend(ctx context.Context) (*Backend, error) {
	wssClient, err := ws.Connect(ctx, rpc.LocalNet.WS)
	if err != nil {
		return nil, err
	}

	backend := &Backend{
		RpcClient: rpc.New(rpc.LocalNet.RPC),
		WssClient: wssClient,
	}

	return backend, nil
}

func (b *Backend) Close() error {
	b.WssClient.Close()
	return b.RpcClient.Close()
}
