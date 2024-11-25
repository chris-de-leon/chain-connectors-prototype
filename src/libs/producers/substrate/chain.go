package substrate

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/chris-de-leon/chain-connectors/src/libs/core"
)

type ChainCursor struct {
	client *gsrpc.SubstrateAPI
}

func NewChainCursor(client *gsrpc.SubstrateAPI) core.Cursor {
	return &ChainCursor{client: client}
}

func NewLogger() *log.Logger {
	return log.New(os.Stdout, fmt.Sprintf("[%s] ", "substrate-block-producer"), log.LstdFlags)
}

func (streamer *ChainCursor) Subscribe(ctx context.Context, cb func(cursor *big.Int)) error {
	sub, err := streamer.client.RPC.Chain.SubscribeFinalizedHeads()
	if err != nil {
		return err
	} else {
		defer sub.Unsubscribe()
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case err, ok := <-sub.Err():
			if !ok {
				return nil
			} else {
				return err
			}
		case data, ok := <-sub.Chan():
			if !ok {
				return nil
			} else {
				cb(new(big.Int).SetUint64(uint64(data.Number)))
			}
		}
	}
}

func (streamer *ChainCursor) GetLatestValue(ctx context.Context) (*big.Int, error) {
	if latestBlock, err := streamer.client.RPC.Chain.GetBlockLatest(); err != nil {
		return nil, err
	} else {
		return new(big.Int).SetUint64(uint64(latestBlock.Block.Header.Number)), nil
	}
}
