package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"

	"github.com/chris-de-leon/chain-connectors/src/libs/common"
	"github.com/chris-de-leon/chain-connectors/src/libs/core"
	"github.com/chris-de-leon/chain-connectors/src/libs/producers/eth"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

func main() {
	serverUrl := common.NewEnvVar("SERVER_URL").AssertExists().AssertNonEmpty()
	wssUrl := common.NewEnvVar("WSS_URL").AssertExists().AssertNonEmpty()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	lis, err := (&net.ListenConfig{}).Listen(ctx, "tcp", serverUrl)
	if err != nil {
		log.Fatal(err)
	} else {
		defer lis.Close()
	}

	wss, err := ethclient.Dial(wssUrl)
	if err != nil {
		log.Fatal(err)
	} else {
		defer wss.Close()
	}

	producer := core.NewProducer(
		grpc.NewServer(),
		core.NewStreamer(
			eth.NewChainCursor(wss),
			eth.NewLogger(),
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
