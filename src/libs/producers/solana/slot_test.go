package solana

import (
	"context"
	"io"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/chris-de-leon/chain-connectors/src/libs/proto"
	"github.com/chris-de-leon/chain-connectors/src/libs/testutils/solana_testutils"
	"github.com/gagliardetto/solana-go/rpc"
	"golang.org/x/net/nettest"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	TESTS_DUR = time.Millisecond * 1100
	TXN_DELAY = time.Millisecond * 100
	TXN_COUNT = 1
)

func TestSolanaBlockServices(t *testing.T) {
	cursorsReceived := []*proto.Cursor{}
	ctx := context.Background()
	eg := new(errgroup.Group)

	backend, err := solana_testutils.InitBackend(ctx, nil)
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

	stm := NewSlotStreamer(backend.RpcClient, backend.WssClient, NewSlotStreamerLogger())
	prd := NewSlotProducer(stm)
	srv := prd.RegisterToServer(grpc.NewServer())

	testCtx, testCancel := context.WithTimeout(ctx, TESTS_DUR)
	defer testCancel()

	eg.Go(func() error {
		return stm.Subscribe(testCtx)
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

	latestSlotNum, err := backend.RpcClient.GetSlot(ctx, rpc.CommitmentFinalized)
	if err != nil {
		t.Fatal(err)
	}

	consumerSlotNum := cursorsReceived[len(cursorsReceived)-1].Value
	if consumerSlotNum != strconv.FormatUint(latestSlotNum, 10) {
		t.Fatalf("consumer did not receive the latest slot (consumer = %s, latest = %d)", consumerSlotNum, latestSlotNum)
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
