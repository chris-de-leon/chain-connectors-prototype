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
	"github.com/chris-de-leon/chain-connectors-prototype/src/plugins/libs/cursor/eth"
	"github.com/chris-de-leon/chain-connectors-prototype/src/plugins/libs/streamer"
	"github.com/ethereum/go-ethereum/ethclient"
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

	client, err := ethclient.Dial(conf.Conn.Wss)
	if err != nil {
		log.Fatal(err)
	} else {
		defer client.Close()
	}

	app := api.New(
		grpc.NewServer(),
		streamer.New(
			eth.NewChainCursor(client),
			eth.NewLogger(),
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
