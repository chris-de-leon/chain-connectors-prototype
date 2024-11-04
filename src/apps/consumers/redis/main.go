package main

import (
	"context"
	"io"
	"log"
	"os/signal"
	"syscall"

	"github.com/chris-de-leon/chain-connectors/src/libs/common"
	"github.com/chris-de-leon/chain-connectors/src/libs/consumers/redis"
	"github.com/chris-de-leon/chain-connectors/src/libs/proto"
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

	redisConsmr := redis.NewBlockConsumer(consumerName, redisClient, redis.NewBlockConsumerLogger(consumerName))
	cursor, err := redisConsmr.Cursor(ctx)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := grpc.NewClient(serverUrl, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}

	stream, err := proto.NewBlockProducerClient(conn).Blocks(ctx, &proto.InitBlock{Height: cursor})
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
				if err := redisConsmr.Consume(ctx, block); err != nil {
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
