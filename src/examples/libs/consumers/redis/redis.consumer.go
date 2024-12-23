package redis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/chris-de-leon/chain-connectors-prototype/proto/go/pb"
	"github.com/redis/go-redis/v9"
)

type Consumer struct {
	client    *redis.Client
	logger    *log.Logger
	streamKey string
	cursorKey string
}

func NewConsumer(name string, client *redis.Client, logger *log.Logger) *Consumer {
	return &Consumer{
		streamKey: fmt.Sprintf("%s:stream", name),
		cursorKey: fmt.Sprintf("%s:cursor", name),
		logger:    logger,
		client:    client,
	}
}

func NewLogger(name string) *log.Logger {
	return log.New(os.Stdout, fmt.Sprintf("[%s-consumer] ", name), log.LstdFlags)
}

func (consumer *Consumer) Cursor(ctx context.Context) (*string, error) {
	cursor, err := consumer.client.HGet(ctx, consumer.cursorKey, "cursor").Result()
	if errors.Is(redis.Nil, err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cursor, nil
}

func (consumer *Consumer) Consume(ctx context.Context, cursor *pb.Cursor) error {
	consumeScript := redis.NewScript(`
    local cursor_key = KEYS[1]
    local stream_key = KEYS[2]

    redis.call("HSET", cursor_key, "cursor", ARGV[1])
    for i = 2, #ARGV do
      redis.call("XADD", stream_key, "*", "data", ARGV[i])
    end

    return ARGV[1]
  `)

	err := consumeScript.Run(
		ctx,
		consumer.client,
		[]string{consumer.cursorKey, consumer.streamKey},
		[]any{cursor.Value, cursor.Value},
	).Err()

	if err != nil {
		return err
	} else {
		consumer.logger.Printf("Cursor: %s", cursor.Value)
	}

	return nil
}
