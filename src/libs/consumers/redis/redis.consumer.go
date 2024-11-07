package redis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/chris-de-leon/chain-connectors/src/libs/proto"
	"github.com/redis/go-redis/v9"
)

type BlockConsumer struct {
	client    *redis.Client
	logger    *log.Logger
	streamKey string
	cursorKey string
}

func NewBlockConsumer(name string, client *redis.Client, logger *log.Logger) *BlockConsumer {
	return &BlockConsumer{
		streamKey: fmt.Sprintf("%s:block-stream", name),
		cursorKey: fmt.Sprintf("%s:block-cursor", name),
		logger:    logger,
		client:    client,
	}
}

func NewBlockConsumerLogger(name string) *log.Logger {
	return log.New(os.Stdout, fmt.Sprintf("[%s-consumer] ", name), log.LstdFlags)
}

func (consumer *BlockConsumer) Cursor(ctx context.Context) (*string, error) {
	cursor, err := consumer.client.HGet(ctx, consumer.cursorKey, "cursor").Result()
	if errors.Is(redis.Nil, err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &cursor, nil
}

func (consumer *BlockConsumer) Consume(ctx context.Context, block *proto.Block) error {
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
		[]any{block.Height, block.Height},
	).Err()

	if err != nil {
		return err
	} else {
		consumer.logger.Printf("Consumed block %s", block.Height)
	}

	return nil
}
