package flow

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/big"
	"os"

	"github.com/chris-de-leon/chain-connectors/src/plugins/libs/cursor"
	"github.com/onflow/flow/protobuf/go/flow/access"
	"github.com/onflow/flow/protobuf/go/flow/executiondata"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChainCursor struct {
	executiondataClient executiondata.ExecutionDataAPIClient
	accessClient        access.AccessAPIClient
}

func NewChainCursor(executiondataClient executiondata.ExecutionDataAPIClient, accessClient access.AccessAPIClient) cursor.Cursor {
	return &ChainCursor{executiondataClient: executiondataClient, accessClient: accessClient}
}

func NewLogger() *log.Logger {
	return log.New(os.Stdout, fmt.Sprintf("[%s] ", "flow-block-cursor"), log.LstdFlags)
}

func (streamer *ChainCursor) Subscribe(ctx context.Context, cb func(cursor *big.Int)) error {
	latestBlock, err := streamer.accessClient.GetLatestBlock(ctx, &access.GetLatestBlockRequest{
		IsSealed: true,
	})
	if err != nil {
		return err
	}

	stream, err := streamer.executiondataClient.SubscribeExecutionDataFromStartBlockHeight(ctx, &executiondata.SubscribeExecutionDataFromStartBlockHeightRequest{
		StartBlockHeight: latestBlock.Block.Height,
	})
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			data, err := stream.Recv()
			if status.Code(err) == codes.Canceled {
				return nil
			}
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			cb(new(big.Int).SetUint64(data.BlockHeight))
		}
	}
}

func (streamer *ChainCursor) GetLatestValue(ctx context.Context) (*big.Int, error) {
	if latestBlockHeader, err := streamer.accessClient.GetLatestBlockHeader(ctx, &access.GetLatestBlockHeaderRequest{IsSealed: true}); err != nil {
		return nil, err
	} else {
		return new(big.Int).SetUint64(latestBlockHeader.Block.Height), nil
	}
}
