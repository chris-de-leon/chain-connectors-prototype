package main

import (
	"context"
	"io"
	"log"
	"math/big"
	"os/signal"
	"syscall"

	"github.com/chris-de-leon/chain-connectors-prototype/src/examples/libs/common"
	"github.com/chris-de-leon/chain-connectors-prototype/src/examples/libs/consumers/redis"
	"github.com/chris-de-leon/chain-connectors-prototype/proto/go/pb"
	redisV9 "github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func main() {
	consumerName := common.NewEnvVar("CONSUMER_NAME").AssertExists().AssertNonEmpty()
	serverUrl := common.NewEnvVar("SERVER_URL").AssertExists().AssertNonEmpty()
	redisUrl := common.NewEnvVar("REDIS_URL").AssertExists().AssertNonEmpty()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	redisClient := redisV9.NewClient(&redisV9.Options{Addr: redisUrl, ContextTimeoutEnabled: true})
	defer redisClient.Close()

	redisConsmr := redis.NewConsumer(consumerName, redisClient, redis.NewLogger(consumerName))
	cursor, err := redisConsmr.Cursor(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if cursor != nil {
		value, ok := new(big.Int).SetString(*cursor, 10)
		if !ok {
			log.Fatalf("failed to convert cursor '%s' to big int", *cursor)
		} else {
			*cursor = new(big.Int).Add(value, new(big.Int).SetUint64(1)).String()
		}
	}

	conn, err := grpc.NewClient(serverUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	stream, err := pb.NewChainCursorClient(conn).Cursors(ctx, &pb.StartCursor{Value: cursor})
	if err != nil {
		log.Fatal(err)
	}

	eg := new(errgroup.Group)
	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
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
				if err := redisConsmr.Consume(ctx, cursor); err != nil {
					return err
				}
			}
		}
	})

	<-ctx.Done()
	if err := conn.Close(); err != nil {
		log.Fatal(err)
	}
	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
}
