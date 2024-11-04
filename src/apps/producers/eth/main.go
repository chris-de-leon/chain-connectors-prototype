package main

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"

	"github.com/chris-de-leon/chain-connectors/src/libs/common"
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

	lgr := eth.NewBlockStreamerLogger()
	stm := eth.NewBlockStreamer(wss, lgr)
	prd := eth.NewBlockProducer(stm)
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
