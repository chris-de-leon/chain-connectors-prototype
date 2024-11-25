package substrate

import (
	"fmt"
	"math/big"

	"github.com/chris-de-leon/chain-connectors/src/libs/proto"
	"google.golang.org/grpc"
)

type BlockProducer struct {
	proto.UnimplementedChainCursorServer
	stream *BlockStreamer
}

func NewBlockProducer(stream *BlockStreamer) BlockProducer {
	return BlockProducer{
		stream: stream,
	}
}

func (producer BlockProducer) RegisterToServer(srv *grpc.Server) *grpc.Server {
	proto.RegisterChainCursorServer(srv, producer)
	return srv
}

func (producer BlockProducer) Cursors(start *proto.StartCursor, stream grpc.ServerStreamingServer[proto.Cursor]) error {
	ctx := stream.Context()

	var cur *big.Int = nil
	if start.Value != nil {
		cursor, ok := new(big.Int).SetString(start.GetValue(), 10)
		if !ok {
			return fmt.Errorf("failed to convert string '%s' to big int", start.GetValue())
		} else {
			cur = cursor
		}
	}

	for {
		value, err := producer.stream.GetNextBlockHeight(ctx, cur)
		if err != nil {
			return err
		}

		if cur == nil {
			cur = value
		}

		for cur.Cmp(value) != 1 {
			if err = stream.Send(&proto.Cursor{Value: cur.String()}); err != nil {
				return err
			} else {
				cur = new(big.Int).Add(cur, big.NewInt(1))
			}
		}
	}
}
