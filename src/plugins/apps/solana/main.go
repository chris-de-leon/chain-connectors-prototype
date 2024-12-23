package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/chris-de-leon/chain-connectors-prototype/src/cli/libs/config"
	"github.com/chris-de-leon/chain-connectors-prototype/src/plugins/libs/api"
	"github.com/chris-de-leon/chain-connectors-prototype/src/plugins/libs/cursor/solana"
	"github.com/chris-de-leon/chain-connectors-prototype/src/plugins/libs/streamer"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	var arg1 []byte
	if len(os.Args) >= 1 {
		arg1 = []byte(os.Args[1])
	} else {
		arg1 = []byte("{}")
	}

	var conf *config.ChainConfig
	if err := json.Unmarshal(arg1, &conf); err != nil {
		log.Fatal(err)
	}

	lis, err := (&net.ListenConfig{}).Listen(ctx, "tcp", conf.Server.Url())

	if err != nil {
		log.Fatal(err)
	} else {
		defer lis.Close()
	}

	wssClient, err := ws.Connect(ctx, conf.Conn.Wss)
	if err != nil {
		log.Fatal(err)
	} else {
		defer wssClient.Close()
	}

	rpcClient := rpc.New(conf.Conn.Rpc)
	defer rpcClient.Close()

	app := api.New(
		grpc.NewServer(),
		streamer.New(
			solana.NewChainCursor(rpcClient, wssClient),
			solana.NewLogger(),
		),
	)

	eg := new(errgroup.Group)
	eg.Go(func() error {
		return app.Stream.Subscribe(ctx)
	})
	eg.Go(func() error {
		return app.Server.Serve(lis)
	})

	log.Printf("Listening on %s\n", conf.Server.Url())
	<-ctx.Done()

	app.Server.GracefulStop()
	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}
