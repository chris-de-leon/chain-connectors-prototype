package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/chris-de-leon/chain-connectors/src/libs/consumers/redis"
	"github.com/chris-de-leon/chain-connectors/src/libs/proto"
	redisV9 "github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	name, exists := os.LookupEnv("CONSUMER_NAME")
	if !exists {
		log.Fatal("env variable 'CONSUMER_NAME' must be set")
	}
	if name == "" {
		log.Fatal("env variable 'CONSUMER_NAME' must not be empty")
	}

	redisClient := redisV9.NewClient(&redisV9.Options{Addr: os.Getenv("REDIS_URL"), ContextTimeoutEnabled: true})
	redisLogger := log.New(os.Stdout, fmt.Sprintf("[%s-consumer] ", name), log.LstdFlags)
	redisConsmr := redis.NewBlockConsumer(name, redisClient, redisLogger)
	cursor, err := redisConsmr.Cursor(ctx)
	if err != nil {
		panic(err)
	}

	conn, err := grpc.NewClient(os.Getenv("SERVER_URL"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	stream, err := proto.NewBlockProducerClient(conn).Blocks(ctx, &proto.InitBlock{Height: cursor})
	if err != nil {
		panic(err)
	}

	eg := new(errgroup.Group)
	eg.Go(func() error {
		for {
			select {
			case <-ctx.Done():
				return nil
			default:
				block, err := stream.Recv()
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
	conn.Close()
	if err := eg.Wait(); err != nil {
		panic(err)
	}
}
