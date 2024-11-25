package flow

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/chris-de-leon/chain-connectors/src/libs/core"
	"github.com/onflow/flow-go-sdk/access/grpc"
)

type ChainCursor struct {
	client *grpc.Client
}

func NewChainCursor(client *grpc.Client) core.Cursor {
	return &ChainCursor{client: client}
}

func NewLogger() *log.Logger {
	return log.New(os.Stdout, fmt.Sprintf("[%s] ", "flow-block-producer"), log.LstdFlags)
}

func (streamer *ChainCursor) Subscribe(ctx context.Context, cb func(cursor *big.Int)) error {
	latestBlock, err := streamer.client.GetLatestBlock(ctx, true)
	if err != nil {
		return err
	}

	dataChan, errChan, err := streamer.client.SubscribeExecutionDataByBlockHeight(ctx, latestBlock.Height)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case err, ok := <-errChan:
			if !ok {
				return nil
			} else {
				return err
			}
		case data, ok := <-dataChan:
			if !ok {
				return nil
			} else {
				cb(new(big.Int).SetUint64(data.Height))
			}
		}
	}
}

func (streamer *ChainCursor) GetLatestValue(ctx context.Context) (*big.Int, error) {
	if latestBlockHeader, err := streamer.client.GetLatestBlockHeader(ctx, true); err != nil {
		return nil, err
	} else {
		return new(big.Int).SetUint64(latestBlockHeader.Height), nil
	}
}
