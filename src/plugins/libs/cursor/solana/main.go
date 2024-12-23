package solana

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/chris-de-leon/chain-connectors-prototype/src/plugins/libs/cursor"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

type ChainCursor struct {
	rpcClient *rpc.Client
	wssClient *ws.Client
	txVersion *uint64
}

func NewChainCursor(rpcClient *rpc.Client, wssClient *ws.Client) cursor.Cursor {
	txVersion := uint64(0)
	return &ChainCursor{
		rpcClient: rpcClient,
		wssClient: wssClient,
		txVersion: &txVersion,
	}
}

func NewLogger() *log.Logger {
	return log.New(os.Stdout, fmt.Sprintf("[%s] ", "solana-slot-cursor"), log.LstdFlags)
}

func (streamer *ChainCursor) Subscribe(ctx context.Context, cb func(cursor *big.Int)) error {
	sub, err := streamer.wssClient.SlotSubscribe()
	if err != nil {
		return err
	} else {
		defer sub.Unsubscribe()
	}

	var lastSlot *uint64 = nil
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
		case _, ok := <-sub.Response():
			if !ok {
				return nil
			}

			slot, err := streamer.rpcClient.GetSlot(ctx, rpc.CommitmentFinalized)
			if ctx.Err() != nil {
				return nil
			}
			if err != nil {
				return err
			}

			if lastSlot == nil || *lastSlot < slot {
				cb(new(big.Int).SetUint64(slot))
			}
			if lastSlot == nil {
				lastSlot = new(uint64)
			}
			*lastSlot = slot
		}
	}
}

func (streamer *ChainCursor) GetLatestValue(ctx context.Context) (*big.Int, error) {
	if latestSlotUint64, err := streamer.rpcClient.GetSlot(ctx, rpc.CommitmentFinalized); err != nil {
		return nil, err
	} else {
		return new(big.Int).SetUint64(latestSlotUint64), nil
	}
}
