package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/chris-de-leon/chain-connectors/src/libs/common"
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

	lgr := substrate.NewBlockStreamerLogger()
	stm := substrate.NewBlockStreamer(rpcClient, lgr)
	prd := substrate.NewBlockProducer(stm)
	srv := prd.RegisterToServer(grpc.NewServer())

	eg := new(errgroup.Group)
	eg.Go(func() error {
		return stm.Subscribe(ctx)
	})
	eg.Go(func() error {
		return srv.Serve(lis)
	})

	lgr.Printf("Listening on %s", serverUrl)
	<-ctx.Done()
	srv.GracefulStop()
	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}
