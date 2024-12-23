package flow

import (
	"context"
	"testing"
	"time"

	"github.com/chris-de-leon/chain-connectors-prototype/src/plugins/libs/api"
	"github.com/chris-de-leon/chain-connectors-prototype/src/plugins/libs/streamer"
	"github.com/chris-de-leon/chain-connectors-prototype/src/plugins/libs/testutils/consumer_testutils"
	"github.com/chris-de-leon/chain-connectors-prototype/src/plugins/libs/testutils/flow_testutils"
	onflow "github.com/onflow/flow-go-sdk/access/grpc"
	"github.com/onflow/flow/protobuf/go/flow/access"
	"github.com/onflow/flow/protobuf/go/flow/executiondata"
	"golang.org/x/net/nettest"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	TESTS_DUR = time.Millisecond * 3750
	TXN_DELAY = time.Millisecond * 100
	TXN_COUNT = 1
)

func TestFlow(t *testing.T) {
	mockConsumer := consumer_testutils.NewChainCursorConsumer()
	ctx := context.Background()
	eg := new(errgroup.Group)

	// NOTE: the gRPC server will automatically close the listener
	lis, err := nettest.NewLocalListener("tcp")
	if err != nil {
		t.Fatal(err)
	}

	acct, err := flow_testutils.NewEmulatorAccount()
	if err != nil {
		t.Fatal(err)
	}

	backend, err := flow_testutils.InitBackend()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Cleanup(func() {
			if err := backend.Close(); err != nil {
				t.Log(err)
			}
		})
	}

	conn, err := grpc.NewClient(onflow.EmulatorHost, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	} else {
		t.Cleanup(func() {
			if err := conn.Close(); err != nil {
				t.Log(err)
			}
		})
	}

	app := api.New(
		grpc.NewServer(),
		streamer.New(
			NewChainCursor(
				executiondata.NewExecutionDataAPIClient(conn),
				access.NewAccessAPIClient(conn),
			),
			NewLogger(),
		),
	)

	testCtx, testCancel := context.WithTimeout(ctx, TESTS_DUR)
	defer testCancel()

	eg.Go(func() error {
		return flow_testutils.
			NewTransactionGenerator(
				acct.SetBackend(backend),
				flow_testutils.NewTransactionGeneratorLogger(),
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
	if err := eg.Wait(); err != nil && status.Code(err) != codes.DeadlineExceeded {
		t.Fatal(err)
	}

	latestBlock, err := backend.GetLatestBlock(ctx, true)
	if err != nil {
		t.Fatal(err)
	}

	mockConsumer.AssertCursorsNotEmpty(t)
	mockConsumer.AssertCursorsInSync(t, latestBlock.Height)
	mockConsumer.AssertCursorsInOrder(t)
}
