package flow

import (
	"context"
	"io"
	"math/big"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chris-de-leon/chain-connectors/src/libs/proto"
	"github.com/chris-de-leon/chain-connectors/src/libs/testutils/flow_testutils"
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

func TestFlowBlockServices(t *testing.T) {
	cursorsReceived := []*proto.Cursor{}
	ctx := context.Background()
	eg := new(errgroup.Group)

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

	lis, err := nettest.NewLocalListener("tcp")
	if err != nil {
		t.Fatal(err)
	} else {
		// NOTE: the gRPC server will automatically close the listener
	}

	gen := flow_testutils.NewTransactionGenerator(acct.SetBackend(backend), flow_testutils.NewTransactionGeneratorLogger())
	stm := NewBlockStreamer(backend, NewBlockStreamerLogger())
	srv := NewBlockProducer(stm).RegisterToServer(grpc.NewServer())

	testCtx, testCancel := context.WithTimeout(ctx, TESTS_DUR)
	defer testCancel()

	eg.Go(func() error {
		if err := gen.Start(testCtx, TXN_DELAY, TXN_COUNT); err != nil {
			if status.Code(err) == codes.DeadlineExceeded {
				return nil
			} else {
				return err
			}
		} else {
			return nil
		}
	})
	eg.Go(func() error {
		if err := stm.Subscribe(testCtx); err != nil {
			if status.Code(err) == codes.Unknown && strings.Contains(err.Error(), "streamer has been stopped") {
				return nil
			} else {
				return err
			}
		} else {
			return nil
		}
	})
	eg.Go(func() error {
		return srv.Serve(lis)
	})

	// NOTE: we can only establish a connection once the server has been started
	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	stream, err := proto.NewChainCursorClient(conn).Cursors(ctx, &proto.StartCursor{Value: nil})
	if err != nil {
		t.Fatal(err)
	}
	eg.Go(func() error {
		for {
			select {
			case <-testCtx.Done():
				return nil
			default:
				cursor, err := stream.Recv()
				if status.Code(err) == codes.Canceled {
					return nil
				}
				if err == io.EOF {
					return nil
				}
				if err != nil {
					return err
				}
				cursorsReceived = append(cursorsReceived, cursor)
			}
		}
	})

	<-testCtx.Done()
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	} else {
		srv.GracefulStop()
	}

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
	if len(cursorsReceived) == 0 {
		t.Fatal("consumer received no blocks")
	}

	latestBlock, err := backend.GetLatestBlock(ctx, true)
	if err != nil {
		t.Fatal(err)
	}

	consumerBlockNum := cursorsReceived[len(cursorsReceived)-1].Value
	if consumerBlockNum != strconv.FormatUint(latestBlock.Height, 10) {
		t.Fatalf("consumer did not receive the latest block (consumer = %s, latest = %d)", consumerBlockNum, latestBlock.Height)
	}

	for i := range len(cursorsReceived) - 1 {
		next := cursorsReceived[i+1].Value
		curr := cursorsReceived[i].Value

		nextHeight, prevOk := new(big.Int).SetString(next, 10)
		if !prevOk {
			t.Fatalf("failed to convert '%s' to big int", next)
		}

		currHeight, currOk := new(big.Int).SetString(curr, 10)
		if !currOk {
			t.Fatalf("failed to convert '%s' to big int", curr)
		}

		if nextHeight.Cmp(new(big.Int).Add(currHeight, new(big.Int).SetUint64(1))) != 0 {
			cursors := make([]string, len(cursorsReceived))
			for i := range cursorsReceived {
				cursors[i] = cursorsReceived[i].Value
			}
			t.Fatalf("Cursors were not received in order: %v", cursors)
		}
	}
}
