package solana_testutils

import (
	"context"
	"fmt"
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

func InitBackend(ctx context.Context, acct *Account) (*Backend, error) {
	wssClient, err := ws.Connect(ctx, rpc.LocalNet.WS)
	if err != nil {
		return nil, err
	}

	backend := &Backend{
		RpcClient: rpc.New(rpc.LocalNet.RPC),
		WssClient: wssClient,
	}

	if acct != nil {
		pubKey := acct.PrivateKey.PublicKey()

		_, err = acct.SetBackend(backend).FundAccount(ctx, uint64(DefaultInitFundAmountInSol))
		if err != nil {
			return nil, err
		}

		bal, err := backend.RpcClient.GetBalance(ctx, pubKey, rpc.CommitmentFinalized)
		if err != nil {
			return nil, err
		}

		fmt.Printf("Account '%s' has been funded with '%d' SOL\n", pubKey.String(), bal.Value)
	}

	return backend, nil
}

func (b *Backend) Close() error {
	b.WssClient.Close()
	return b.RpcClient.Close()
}
