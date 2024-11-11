package solana

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"sync"

	"github.com/gagliardetto/solana-go/rpc"
	"github.com/gagliardetto/solana-go/rpc/ws"
)

var ErrStreamerStopped = errors.New("streamer has been stopped")

type SlotStreamer struct {
	rpcClient *rpc.Client
	wssClient *ws.Client
	logger    *log.Logger
	signal    *sync.Cond
	txVersion *uint64
	isStopped bool
}

func NewSlotStreamer(rpcClient *rpc.Client, wssClient *ws.Client, logger *log.Logger) *SlotStreamer {
	txVersion := uint64(0)
	return &SlotStreamer{
		signal:    sync.NewCond(&sync.Mutex{}),
		rpcClient: rpcClient,
		wssClient: wssClient,
		logger:    logger,
		txVersion: &txVersion,
		isStopped: false,
	}
}

func NewSlotStreamerLogger() *log.Logger {
	return log.New(os.Stdout, fmt.Sprintf("[%s] ", "solana-slot-producer"), log.LstdFlags)
}

func (streamer *SlotStreamer) Subscribe(ctx context.Context) error {
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

	sub, err := streamer.wssClient.SlotSubscribe()
	if err != nil {
		return err
	}

	streamer.logger.Printf("Waiting for new finalized slots...")
	var lastSlot *uint64 = nil

	for {
		select {
		case <-ctx.Done():
			sub.Unsubscribe()
			return nil
		case err, ok := <-sub.Err():
			if !ok {
				return nil
			} else {
				return err
			}
		case _, ok := <-sub.Response():
			if !ok {
				return nil
			}

			slot, err := streamer.rpcClient.GetSlot(ctx, rpc.CommitmentFinalized)
			if ctx.Err() != nil {
				return nil
			}
			if err != nil {
				return err
			}

			if lastSlot == nil || *lastSlot < slot {
				streamer.logger.Printf("New finalized slot: %d", slot)
				streamer.signal.Broadcast()
			}
			if lastSlot == nil {
				lastSlot = new(uint64)
			}
			*lastSlot = slot
		}
	}
}

func (streamer *SlotStreamer) WaitForNextFinalizedSlot(ctx context.Context) error {
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

func (streamer *SlotStreamer) GetNextSlot(ctx context.Context, curr *big.Int) (*big.Int, error) {
	if streamer.isStopped {
		return nil, ErrStreamerStopped
	}

	latestSlotUint64, err := streamer.rpcClient.GetSlot(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}

	latestSlotBigInt := new(big.Int).SetUint64(latestSlotUint64)
	if curr == nil || curr.Cmp(latestSlotBigInt) == -1 {
		return latestSlotBigInt, nil
	}

	if err := streamer.WaitForNextFinalizedSlot(ctx); err != nil {
		return nil, err
	} else {
		return new(big.Int).Add(latestSlotBigInt, new(big.Int).SetUint64(1)), nil
	}
}
