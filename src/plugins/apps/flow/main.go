package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/chris-de-leon/chain-connectors/src/cli/libs/config"
	"github.com/chris-de-leon/chain-connectors/src/plugins/libs/api"
	"github.com/chris-de-leon/chain-connectors/src/plugins/libs/cursor/flow"
	"github.com/chris-de-leon/chain-connectors/src/plugins/libs/streamer"
	"github.com/onflow/flow/protobuf/go/flow/access"
	"github.com/onflow/flow/protobuf/go/flow/executiondata"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	conn, err := grpc.NewClient(conf.Conn.Wss, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	} else {
		defer conn.Close()
	}

	app := api.New(
		grpc.NewServer(),
		streamer.New(
			flow.NewChainCursor(
				executiondata.NewExecutionDataAPIClient(conn),
				access.NewAccessAPIClient(conn),
			),
			flow.NewLogger(),
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
