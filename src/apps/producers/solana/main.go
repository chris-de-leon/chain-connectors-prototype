package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"

	"github.com/chris-de-leon/chain-connectors/src/libs/common"
	"github.com/chris-de-leon/chain-connectors/src/libs/core"
	"github.com/chris-de-leon/chain-connectors/src/libs/producers/solana"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

func main() {
	serverUrl := common.NewEnvVar("SERVER_URL").AssertExists().AssertNonEmpty()
	rpcUrl := common.NewEnvVar("RPC_URL").AssertExists().AssertNonEmpty()
	wssUrl := common.NewEnvVar("WSS_URL").AssertExists().AssertNonEmpty()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	lis, err := (&net.ListenConfig{}).Listen(ctx, "tcp", serverUrl)
	if err != nil {
		log.Fatal(err)
	} else {
		defer lis.Close()
	}

	wssClient, err := ws.Connect(ctx, wssUrl)
	if err != nil {
		log.Fatal(err)
	} else {
		defer wssClient.Close()
	}

	rpcClient := rpc.New(rpcUrl)
	defer rpcClient.Close()

	producer := core.NewProducer(
		grpc.NewServer(),
		core.NewStreamer(
			solana.NewChainCursor(rpcClient, wssClient),
			solana.NewLogger(),
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
