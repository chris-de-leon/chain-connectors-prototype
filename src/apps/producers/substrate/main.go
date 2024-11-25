package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/chris-de-leon/chain-connectors/src/libs/common"
	"github.com/chris-de-leon/chain-connectors/src/libs/core"
	"github.com/chris-de-leon/chain-connectors/src/libs/producers/substrate"
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

	rpcClient, err := gsrpc.NewSubstrateAPI(rpcUrl)
	if err != nil {
		log.Fatal(err)
	} else {
		defer rpcClient.Client.Close()
	}

	producer := core.NewProducer(
		grpc.NewServer(),
		core.NewStreamer(
			substrate.NewChainCursor(rpcClient),
			substrate.NewLogger(),
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
