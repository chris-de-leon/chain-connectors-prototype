package eth

import (
	"context"
	"testing"
	"time"

	"github.com/chris-de-leon/chain-connectors/src/plugins/libs/api"
	"github.com/chris-de-leon/chain-connectors/src/plugins/libs/streamer"
	"github.com/chris-de-leon/chain-connectors/src/plugins/libs/testutils/consumer_testutils"
	"github.com/chris-de-leon/chain-connectors/src/plugins/libs/testutils/eth_testutils"
	"golang.org/x/net/nettest"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const (
	TESTS_DUR = time.Millisecond * 3750
	TXN_DELAY = time.Millisecond * 100
	TXN_COUNT = 1
)

func TestEth(t *testing.T) {
	mockConsumer := consumer_testutils.NewChainCursorConsumer()
	ctx := context.Background()
	eg := new(errgroup.Group)

	// NOTE: the gRPC server will automatically close the listener
	lis, err := nettest.NewLocalListener("tcp")
	if err != nil {
		t.Fatal(err)
	}

	acct, err := eth_testutils.NewAccount()
	if err != nil {
		t.Fatal(err)
	}

	backend, err := eth_testutils.InitBackend(acct)
	if err != nil {
		t.Fatal(err)
	} else {
		t.Cleanup(func() {
			if err := backend.Close(); err != nil {
				t.Log(err)
			}
		})
	}

	app := api.New(
		grpc.NewServer(),
		streamer.New(
			NewChainCursor(backend.Client()),
			NewLogger(),
		),
	)

	testCtx, testCancel := context.WithTimeout(ctx, TESTS_DUR)
	defer testCancel()

	eg.Go(func() error {
		return eth_testutils.
			NewTransactionGenerator(
				acct.SetBackend(backend),
				eth_testutils.NewTransactionGeneratorLogger(),
			).
			Start(testCtx, TXN_DELAY, TXN_COUNT)
	})
	eg.Go(func() error {
		return app.Stream.Subscribe(testCtx)
	})
	eg.Go(func() error {
		return app.Server.Serve(lis)
	})
	eg.Go(func() error {
		return mockConsumer.Listen(testCtx, lis.Addr().String())
	})

	<-testCtx.Done()
	if err := mockConsumer.Close(); err != nil {
		t.Fatal(err)
	}

	app.Server.GracefulStop()
	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}

	latestBlockNum, err := backend.Client().BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}

	mockConsumer.AssertCursorsNotEmpty(t)
	mockConsumer.AssertCursorsInSync(t, latestBlockNum)
	mockConsumer.AssertCursorsInOrder(t)
}
