package substrate

import (
	"context"
	"io"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/chris-de-leon/chain-connectors/src/libs/core"
	"github.com/chris-de-leon/chain-connectors/src/libs/proto"
	"github.com/chris-de-leon/chain-connectors/src/libs/testutils/substrate_testutils"
	"golang.org/x/net/nettest"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	TESTS_DUR = time.Millisecond * 6750
	TXN_DELAY = time.Millisecond * 100
	TXN_COUNT = 1
)

func TestSubstrate(t *testing.T) {
	cursorsReceived := []*proto.Cursor{}
	ctx := context.Background()
	eg := new(errgroup.Group)

	backend, err := substrate_testutils.InitBackend()
	if err != nil {
		t.Fatal(err)
	} else {
		t.Cleanup(func() {
			backend.Client.Close()
		})
	}

	lis, err := nettest.NewLocalListener("tcp")
	if err != nil {
		t.Fatal(err)
	} else {
		// NOTE: the gRPC server will automatically close the listener
	}

	prd := core.NewProducer(
		grpc.NewServer(),
		core.NewStreamer(
			NewChainCursor(backend),
			NewLogger(),
		),
	)

	testCtx, testCancel := context.WithTimeout(ctx, TESTS_DUR)
	defer testCancel()

	eg.Go(func() error {
		return prd.Stream.Subscribe(testCtx)
	})
	eg.Go(func() error {
		return prd.Server.Serve(lis)
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
		prd.Server.GracefulStop()
	}

	if err := eg.Wait(); err != nil {
		t.Fatal(err)
	}
	if len(cursorsReceived) == 0 {
		t.Fatal("consumer received no blocks")
	}

	latestblock, err := backend.RPC.Chain.GetBlockLatest()
	if err != nil {
		t.Fatal(err)
	}

	consumerSlotNum := cursorsReceived[len(cursorsReceived)-1].Value
	if consumerSlotNum != strconv.FormatUint(uint64(latestblock.Block.Header.Number), 10) {
		t.Fatalf("consumer did not receive the latest block (consumer = %s, latest = %d)", consumerSlotNum, latestblock.Block.Header.Number)
	}

	for i := range len(cursorsReceived) - 1 {
		next := cursorsReceived[i+1].Value
		curr := cursorsReceived[i].Value

		nextSlot, prevOk := new(big.Int).SetString(next, 10)
		if !prevOk {
			t.Fatalf("failed to convert '%s' to big int", next)
		}

		currSlot, currOk := new(big.Int).SetString(curr, 10)
		if !currOk {
			t.Fatalf("failed to convert '%s' to big int", curr)
		}

		if nextSlot.Cmp(new(big.Int).Add(currSlot, new(big.Int).SetUint64(1))) != 0 {
			cursors := make([]string, len(cursorsReceived))
			for i := range cursorsReceived {
				cursors[i] = cursorsReceived[i].Value
			}
			t.Fatalf("Cursors were not received in order: %v", cursors)
		}
	}
}
