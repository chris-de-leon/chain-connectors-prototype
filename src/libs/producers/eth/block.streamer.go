package eth

import (
	"context"
	"log"
	"sync"

	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type BlockStreamer struct {
	client *ethclient.Client
	logger *log.Logger
	signal *sync.Cond
	mutex  *sync.Mutex
}

func NewBlockStreamer(client *ethclient.Client, logger *log.Logger) *BlockStreamer {
	mutex := &sync.Mutex{}
	return &BlockStreamer{
		signal: sync.NewCond(mutex),
		logger: logger,
		client: client,
		mutex:  mutex,
	}
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
				streamer.mutex.Lock()
				streamer.signal.Broadcast()
				streamer.mutex.Unlock()
				streamer.logger.Printf("Received block %s", header.Number.String())
			}
		}
	}
}

func (streamer *BlockStreamer) WaitForNextBlockHeader(ctx context.Context) {
	streamer.mutex.Lock()
	defer streamer.mutex.Unlock()
	streamer.signal.Wait()
}
