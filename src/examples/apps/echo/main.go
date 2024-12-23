package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/signal"
	"syscall"

	"github.com/chris-de-leon/chain-connectors/src/examples/libs/common"
	"github.com/chris-de-leon/chain-connectors/proto/go/pb"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	serverUrl := common.NewEnvVar("SERVER_URL").AssertExists().AssertNonEmpty()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	conn, err := grpc.NewClient(serverUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	stream, err := pb.NewChainCursorClient(conn).Cursors(ctx, &pb.StartCursor{Value: nil})
	if err != nil {
		log.Fatal(err)
	}

	eg := new(errgroup.Group)
	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				cursor, err := stream.Recv()
				if status.Code(err) == codes.Canceled {
					return nil
				}
				if err == io.EOF {
					return nil
				}
				if err != nil {
					return err
				}
				fmt.Printf("Received cursor %s\n", cursor.Value)
			}
		}
	})

	<-ctx.Done()
	if err := conn.Close(); err != nil {
		log.Fatal(err)
	}
	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}
