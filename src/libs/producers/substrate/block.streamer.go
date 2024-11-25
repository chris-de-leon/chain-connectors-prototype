package substrate

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"sync"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
)

var ErrStreamerStopped = errors.New("streamer has been stopped")

type BlockStreamer struct {
	client    *gsrpc.SubstrateAPI
	logger    *log.Logger
	signal    *sync.Cond
	isStopped bool
}

func NewBlockStreamer(client *gsrpc.SubstrateAPI, logger *log.Logger) *BlockStreamer {
	return &BlockStreamer{
		signal:    sync.NewCond(&sync.Mutex{}),
		client:    client,
		logger:    logger,
		isStopped: false,
	}
}

func NewBlockStreamerLogger() *log.Logger {
	return log.New(os.Stdout, fmt.Sprintf("[%s] ", "substrate-block-producer"), log.LstdFlags)
}

func (streamer *BlockStreamer) Subscribe(ctx context.Context) error {
	if streamer.isStopped {
		return ErrStreamerStopped
	} else {
		defer func() {
			// NOTE: if the context is cancelled, then we need to make sure that (1) any
			// goroutines waiting for this signal are unblocked so that they can free up
			// their resources and exit gracefully and (2) any further attempts to Wait()
			// on the signal are prevented (we no longer intend to call Broadcast() after
			// the context is closed, so if we try to Wait() on this signal after the ctx
			// is cancelled, then this will create dangling goroutines).
			streamer.signal.L.Lock()
			streamer.isStopped = true
			streamer.signal.Broadcast()
			streamer.signal.L.Unlock()
		}()
	}

	sub, err := streamer.client.RPC.Chain.SubscribeFinalizedHeads()
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
		case data, ok := <-sub.Chan():
			if !ok {
				return nil
			} else {
				streamer.signal.Broadcast()
				streamer.logger.Printf("New block: %d", data.Number)
			}
		}
	}
}

func (streamer *BlockStreamer) WaitForNextBlockHeight(ctx context.Context) error {
	if streamer.isStopped {
		return ErrStreamerStopped
	}

	done := make(chan struct{})
	errs := make(chan error)
	defer close(done)
	defer close(errs)

	go func() {
		// NOTE: we need to ensure that the mutex is locked beforehand since Wait() will try
		// to unlock it
		//
		// NOTE: once Wait() unlocks the mutex it will suspend execution - keep in mind that
		// the mutex is **NOT** locked while execution is suspended
		//
		// NOTE: if Broadcast() or Signal() are called, then Wait() will resume where it left
		// off and lock the mutex then return - this is why we unlock the mutex afterwards
		//
		// NOTE: if multiple go routines call this function, then each call to Wait() will be
		// performed atomically
		streamer.signal.L.Lock()
		if !streamer.isStopped {
			streamer.signal.Wait()
		}
		streamer.signal.L.Unlock()

		// NOTE: if the context is done, then we will simply exit the function. In this case,
		// both channels will already be closed, so there is no point in writing data to them
		if ctx.Err() != nil {
			return
		} else {
			if streamer.isStopped {
				errs <- ErrStreamerStopped
			} else {
				done <- struct{}{}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errs:
			return err
		case <-done:
			return nil
		}
	}
}

func (streamer *BlockStreamer) GetNextBlockHeight(ctx context.Context, curr *big.Int) (*big.Int, error) {
	if streamer.isStopped {
		return nil, ErrStreamerStopped
	}

	latestBlock, err := streamer.client.RPC.Chain.GetBlockLatest()
	if err != nil {
		return nil, err
	}

	latestBlockNum := new(big.Int).SetUint64(uint64(latestBlock.Block.Header.Number))
	if curr == nil || curr.Cmp(latestBlockNum) == -1 {
		return latestBlockNum, nil
	}

	if err := streamer.WaitForNextBlockHeight(ctx); err != nil {
		return nil, err
	} else {
		return new(big.Int).Add(latestBlockNum, new(big.Int).SetUint64(1)), nil
	}
}
