package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"

	"github.com/chris-de-leon/chain-connectors/src/libs/common"
	"github.com/chris-de-leon/chain-connectors/src/libs/core"
	"github.com/chris-de-leon/chain-connectors/src/libs/producers/flow"
	onflow "github.com/onflow/flow-go-sdk/access/grpc"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

func main() {
	serverUrl := common.NewEnvVar("SERVER_URL").AssertExists().AssertNonEmpty()
	rpcUrl := common.NewEnvVar("RPC_URL").AssertExists().AssertNonEmpty()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	lis, err := (&net.ListenConfig{}).Listen(ctx, "tcp", serverUrl)
	if err != nil {
		log.Fatal(err)
	} else {
		defer lis.Close()
	}

	client, err := onflow.NewClient(rpcUrl)
	if err != nil {
		log.Fatal(err)
	} else {
		defer client.Close()
	}

	producer := core.NewProducer(
		grpc.NewServer(),
		core.NewStreamer(
			flow.NewChainCursor(client),
			flow.NewLogger(),
		),
	)

	eg := new(errgroup.Group)
	eg.Go(func() error {
		return producer.Stream.Subscribe(ctx)
	})
	eg.Go(func() error {
		return producer.Server.Serve(lis)
	})

	log.Printf("Listening on %s\n", serverUrl)
	<-ctx.Done()
	producer.Server.GracefulStop()
	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}
