package eth

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/chris-de-leon/chain-connectors/src/libs/proto"
	"google.golang.org/grpc"
)

type BlockProducer struct {
	proto.UnimplementedBlockProducerServer
	stream *BlockStreamer
}

func NewBlockProducer(stream *BlockStreamer) BlockProducer {
	return BlockProducer{
		stream: stream,
	}
}

func (producer BlockProducer) RegisterToServer(srv *grpc.Server) *grpc.Server {
	proto.RegisterBlockProducerServer(srv, producer)
	return srv
}

func (producer BlockProducer) Blocks(initBlock *proto.InitBlock, stream grpc.ServerStreamingServer[proto.Block]) error {
	ctx := stream.Context()

	cursor, err := producer.initCursor(ctx, initBlock)
	if err != nil {
		return err
	}

	for {
		height, err := producer.stream.WaitForNextBlockHeight(ctx, cursor)
		if err != nil {
			return err
		}

		err = stream.Send(&proto.Block{Height: height})
		if err != nil {
			return err
		}

		cursor = new(big.Int).Add(cursor, big.NewInt(1))
	}
}

func (producer BlockProducer) initCursor(ctx context.Context, initBlock *proto.InitBlock) (*big.Int, error) {
	latestBlock, err := producer.stream.client.BlockByNumber(ctx, nil)
	if err != nil {
		return nil, err
	}
	if latestBlock == nil {
		return nil, errors.New("failed to retrieve the latest block")
	}
	if initBlock.Height == nil {
		return latestBlock.Number(), nil
	}

	cursor, ok := new(big.Int).SetString(initBlock.GetHeight(), 10)
	if !ok {
		return nil, fmt.Errorf("failed to convert string '%s' to big int", initBlock.GetHeight())
	}
	if cursor.Cmp(big.NewInt(0)) == -1 {
		return nil, errors.New("invalid start block height (must be >= 0)")
	}
	if cursor.Cmp(latestBlock.Number()) == 1 {
		return nil, errors.New("starting block height is larger than latest block height")
	}

	return cursor, nil
}
