package eth

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"sync"

	"github.com/ethereum/go-ethereum"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
)

type BlockReader interface {
	ethereum.BlockNumberReader
	ethereum.ChainReader
}

type BlockStreamer struct {
	client BlockReader
	logger *log.Logger
	signal *sync.Cond
}

func NewBlockStreamer(client BlockReader, logger *log.Logger) *BlockStreamer {
	return &BlockStreamer{
		signal: sync.NewCond(&sync.Mutex{}),
		logger: logger,
		client: client,
	}
}

func NewBlockStreamerLogger() *log.Logger {
	return log.New(os.Stdout, fmt.Sprintf("[%s] ", "eth-block-producer"), log.LstdFlags)
}

func (streamer *BlockStreamer) Subscribe(ctx context.Context) error {
	headers := make(chan *ethtypes.Header)
	defer close(headers)

	sub, err := streamer.client.SubscribeNewHead(ctx, headers)
	if err != nil {
		return err
	} else {
		defer sub.Unsubscribe()
	}

	streamer.logger.Printf("Waiting for new blocks...")
	for {
		select {
		case <-ctx.Done():
			return nil
		case err, ok := <-sub.Err():
			if !ok {
				return nil
			} else {
				return err
			}
		case header, ok := <-headers:
			if !ok {
				return nil
			} else {
				streamer.signal.L.Lock()
				streamer.signal.Broadcast()
				streamer.signal.L.Unlock()
				streamer.logger.Printf("Received block %s", header.Number.String())
			}
		}
	}
}

func (streamer *BlockStreamer) WaitForNextBlockHeader(ctx context.Context) error {
	done := make(chan struct{})
	defer close(done)

	go func() {
		streamer.signal.L.Lock()
		streamer.signal.Wait()
		streamer.signal.L.Unlock()
		done <- struct{}{}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-done:
			return nil
		}
	}
}

func (streamer *BlockStreamer) WaitForNextBlock(ctx context.Context, currHeight *big.Int) (*ethtypes.Block, error) {
	nextHeight := new(big.Int).Add(currHeight, big.NewInt(1))

	latestHeight, err := streamer.client.BlockNumber(ctx)
	if err != nil {
		return nil, err
	}
	if nextHeight.Cmp(new(big.Int).SetUint64(latestHeight)) == 1 {
		if err := streamer.WaitForNextBlockHeader(ctx); err != nil {
			return nil, err
		}
	}

	block, err := streamer.client.BlockByNumber(ctx, nextHeight)
	if err != nil {
		return nil, err
	}
	if block == nil {
		return nil, fmt.Errorf("unexpectedly received empty block for height '%d'", nextHeight)
	}

	return block, nil
}
