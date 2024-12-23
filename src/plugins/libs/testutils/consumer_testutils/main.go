package consumer_testutils

import (
	"context"
	"io"
	"math/big"
	"strconv"
	"testing"

	"github.com/chris-de-leon/chain-connectors-prototype/proto/go/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type (
	Grpc struct {
		Client pb.ChainCursorClient
		Conn   *grpc.ClientConn
	}

	ChainCursorConsumer struct {
		Grpc    *Grpc
		Cursors []*pb.Cursor
	}
)

func NewChainCursorConsumer() ChainCursorConsumer {
	return ChainCursorConsumer{Cursors: []*pb.Cursor{}, Grpc: nil}
}

func (c *ChainCursorConsumer) Close() error {
	return c.Grpc.Conn.Close()
}

func (c *ChainCursorConsumer) Connect(url string) error {
	if c.Grpc != nil {
		return nil
	}

	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	c.Grpc = &Grpc{
		Client: pb.NewChainCursorClient(conn),
		Conn:   conn,
	}

	return nil
}

func (c *ChainCursorConsumer) Listen(ctx context.Context, url string) error {
	if err := c.Connect(url); err != nil {
		return err
	}

	stream, err := c.Grpc.Client.Cursors(ctx, &pb.StartCursor{Value: nil})
	if err != nil {
		return err
	}

	for {
		cursor, err := stream.Recv()
		if status.Code(err) == codes.DeadlineExceeded {
			return nil
		}
		if status.Code(err) == codes.Canceled {
			return nil
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		c.Cursors = append(c.Cursors, cursor)
	}
}

func (c *ChainCursorConsumer) AssertCursorsNotEmpty(t *testing.T) {
	if len(c.Cursors) == 0 {
		t.Fatal("consumer received no blocks")
	}
}

func (c *ChainCursorConsumer) AssertCursorsInSync(t *testing.T, latestCursor uint64) {
	consumerCursor := c.Cursors[len(c.Cursors)-1].Value
	if consumerCursor != strconv.FormatUint(latestCursor, 10) {
		t.Fatalf(
			"consumer did not receive the latest block (consumer = %s, latest = %d)",
			consumerCursor,
			latestCursor,
		)
	}
}

func (c *ChainCursorConsumer) AssertCursorsInOrder(t *testing.T) {
	for i := range len(c.Cursors) - 1 {
		next := c.Cursors[i+1].Value
		curr := c.Cursors[i].Value

		nextHeight, prevOk := new(big.Int).SetString(next, 10)
		if !prevOk {
			t.Fatalf("failed to convert '%s' to big int", next)
		}

		currHeight, currOk := new(big.Int).SetString(curr, 10)
		if !currOk {
			t.Fatalf("failed to convert '%s' to big int", curr)
		}

		if nextHeight.Cmp(new(big.Int).Add(currHeight, new(big.Int).SetUint64(1))) != 0 {
			cursors := make([]string, len(c.Cursors))
			for i := range c.Cursors {
				cursors[i] = c.Cursors[i].Value
			}
			t.Fatalf("Cursors were not received in order: %v", cursors)
		}
	}
}
