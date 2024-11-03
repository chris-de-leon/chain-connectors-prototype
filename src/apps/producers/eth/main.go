package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/chris-de-leon/chain-connectors/src/libs/producers/eth"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	wss, err := ethclient.Dial(os.Getenv("WSS_URL"))
	if err != nil {
		panic(err)
	}

	lis, err := net.Listen("tcp", os.Getenv("SERVER_URL"))
	if err != nil {
		panic(err)
	}

	lgr := log.New(os.Stdout, fmt.Sprintf("[%s] ", "eth-block-producer"), log.LstdFlags)
	stm := eth.NewBlockStreamer(wss, lgr)
	prd := eth.NewBlockProducer(stm)
	srv := grpc.NewServer()

	prd.RegisterToServer(srv)
	eg := new(errgroup.Group)
	eg.Go(func() error {
		return stm.Subscribe(ctx)
	})
	eg.Go(func() error {
		return srv.Serve(lis)
	})

	<-ctx.Done()
	srv.GracefulStop()
	if err := eg.Wait(); err != nil {
		panic(err)
	}
}
