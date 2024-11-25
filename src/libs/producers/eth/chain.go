package eth

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/chris-de-leon/chain-connectors/src/libs/core"
	"github.com/ethereum/go-ethereum"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type BlockReader interface {
	ethereum.BlockNumberReader
	ethereum.ChainReader
}

type ChainCursor struct {
	client BlockReader
}

func NewChainCursor(client BlockReader) core.Cursor {
	return &ChainCursor{client}
}

func NewLogger() *log.Logger {
	return log.New(os.Stdout, fmt.Sprintf("[%s] ", "eth-block-producer"), log.LstdFlags)
}

func (streamer *ChainCursor) Subscribe(ctx context.Context, cb func(cursor *big.Int)) error {
	headers := make(chan *ethtypes.Header)
	defer close(headers)

	sub, err := streamer.client.SubscribeNewHead(ctx, headers)
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
		case header, ok := <-headers:
			if !ok {
				return nil
			} else {
				cb(header.Number)
			}
		}
	}
}

func (streamer *ChainCursor) GetLatestValue(ctx context.Context) (*big.Int, error) {
	if latestBlockNumUint64, err := streamer.client.BlockNumber(ctx); err != nil {
		return nil, err
	} else {
		return new(big.Int).SetUint64(latestBlockNumUint64), nil
	}
}
