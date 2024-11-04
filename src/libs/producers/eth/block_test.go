package eth

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/chris-de-leon/chain-connectors/src/libs/proto"
	"github.com/chris-de-leon/chain-connectors/src/libs/testutils/geth"
	"golang.org/x/net/nettest"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	TESTS_DURATION = time.Duration(1100)
	BLOCK_RATE_DUR = time.Duration(100)
	TXNS_PER_BLOCK = 1
)

func TestEthereumBlockServices(t *testing.T) {
	blocksReceived := []*proto.Block{}
	ctx := context.Background()
	eg := new(errgroup.Group)

	acct, err := geth.NewAccount()
	if err != nil {
		t.Fatal(err)
	}

	backend, err := geth.InitBackend(acct)
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

	gen := geth.NewBlockGenerator(acct.SetBackend(backend), geth.NewBlockGeneratorLogger())
	stm := NewBlockStreamer(backend.Client(), NewBlockStreamerLogger())
	prd := NewBlockProducer(stm)
	srv := prd.RegisterToServer(grpc.NewServer())

	testCtx, testCancel := context.WithTimeout(ctx, TESTS_DURATION*time.Millisecond)
	defer testCancel()

	eg.Go(func() error {
		return gen.Start(testCtx, BLOCK_RATE_DUR, TXNS_PER_BLOCK)
	})
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
	stream, err := proto.NewBlockProducerClient(conn).Blocks(ctx, &proto.InitBlock{Height: nil})
	if err != nil {
		t.Fatal(err)
	}
	eg.Go(func() error {
		for {
			select {
			case <-testCtx.Done():
				return nil
			default:
				block, err := stream.Recv()
				if status.Code(err) == codes.Canceled {
					return nil
				}
				if err == io.EOF {
					return nil
				}
				if err != nil {
					return err
				}
				blocksReceived = append(blocksReceived, block)
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

	blockNum, err := backend.Client().BlockNumber(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if uint64(len(blocksReceived)) != blockNum {
		t.Fatalf("expected to receive '%d' blocks but got '%d'", blockNum, len(blocksReceived))
	}

	for i, b := range blocksReceived {
		if b.Height != fmt.Sprintf("%d", i+1) {
			heights := make([]string, len(blocksReceived))
			for i := range blocksReceived {
				heights[i] = blocksReceived[i].Height
			}
			t.Fatalf("Blocks were not received in order: %v", heights)
		}
	}
}
