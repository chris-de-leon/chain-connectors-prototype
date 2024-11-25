package core

import (
	"fmt"
	"math/big"

	"github.com/chris-de-leon/chain-connectors/src/libs/proto"
	"google.golang.org/grpc"
)

type Producer struct {
	proto.UnimplementedChainCursorServer
	Server *grpc.Server
	Stream *Streamer
}

func NewProducer(server *grpc.Server, stream *Streamer) *Producer {
	producer := &Producer{Server: server, Stream: stream}
	proto.RegisterChainCursorServer(server, producer)
	return producer
}

func (producer *Producer) Cursors(start *proto.StartCursor, stream grpc.ServerStreamingServer[proto.Cursor]) error {
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
		value, err := producer.Stream.GetNextCursor(ctx, cur)
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
