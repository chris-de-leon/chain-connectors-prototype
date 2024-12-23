package substrate

import (
	"context"
	"testing"
	"time"

	"github.com/chris-de-leon/chain-connectors-prototype/src/plugins/libs/api"
	"github.com/chris-de-leon/chain-connectors-prototype/src/plugins/libs/streamer"
	"github.com/chris-de-leon/chain-connectors-prototype/src/plugins/libs/testutils/consumer_testutils"
	"github.com/chris-de-leon/chain-connectors-prototype/src/plugins/libs/testutils/substrate_testutils"
	"golang.org/x/net/nettest"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

const (
	TESTS_DUR = time.Millisecond * 6750
	TXN_DELAY = time.Millisecond * 100
	TXN_COUNT = 1
)

func TestSubstrate(t *testing.T) {
	mockConsumer := consumer_testutils.NewChainCursorConsumer()
	ctx := context.Background()
	eg := new(errgroup.Group)

	// NOTE: the gRPC server will automatically close the listener
	lis, err := nettest.NewLocalListener("tcp")
	if err != nil {
		t.Fatal(err)
	}

	backend, err := substrate_testutils.InitBackend()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Cleanup(func() {
			backend.Client.Close()
		})
	}

	app := api.New(
		grpc.NewServer(),
		streamer.New(
			NewChainCursor(backend),
			NewLogger(),
		),
	)

	testCtx, testCancel := context.WithTimeout(ctx, TESTS_DUR)
	defer testCancel()

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

	latestblock, err := backend.RPC.Chain.GetBlockLatest()
	if err != nil {
		t.Fatal(err)
	}

	mockConsumer.AssertCursorsNotEmpty(t)
	mockConsumer.AssertCursorsInSync(t, uint64(latestblock.Block.Header.Number))
	mockConsumer.AssertCursorsInOrder(t)
}
