package api

import (
	"fmt"
	"math/big"

	"github.com/chris-de-leon/chain-connectors-prototype/proto/go/pb"
	"github.com/chris-de-leon/chain-connectors-prototype/src/plugins/libs/streamer"
	"google.golang.org/grpc"
)

type API struct {
	pb.UnimplementedChainCursorServer
	Server *grpc.Server
	Stream *streamer.Streamer
}

func New(server *grpc.Server, stream *streamer.Streamer) *API {
	api := &API{Server: server, Stream: stream}
	pb.RegisterChainCursorServer(server, api)
	return api
}

func (api *API) Cursors(start *pb.StartCursor, stream grpc.ServerStreamingServer[pb.Cursor]) error {
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
		value, err := api.Stream.GetNextCursor(ctx, cur)
		if err != nil {
			return err
		}

		if cur == nil {
			cur = value
		}

		for cur.Cmp(value) != 1 {
			if err = stream.Send(&pb.Cursor{Value: cur.String()}); err != nil {
				return err
			} else {
				cur = new(big.Int).Add(cur, big.NewInt(1))
			}
		}
	}
}
