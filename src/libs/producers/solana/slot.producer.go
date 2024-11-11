package solana

import (
	"fmt"
	"math/big"

	"github.com/chris-de-leon/chain-connectors/src/libs/proto"
	"google.golang.org/grpc"
)

type SlotProducer struct {
	proto.UnimplementedChainCursorServer
	stream *SlotStreamer
}

func NewSlotProducer(stream *SlotStreamer) SlotProducer {
	return SlotProducer{
		stream: stream,
	}
}

func (producer SlotProducer) RegisterToServer(srv *grpc.Server) *grpc.Server {
	proto.RegisterChainCursorServer(srv, producer)
	return srv
}

func (producer SlotProducer) Cursors(start *proto.StartCursor, stream grpc.ServerStreamingServer[proto.Cursor]) error {
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
		value, err := producer.stream.GetNextSlot(ctx, cur)
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
