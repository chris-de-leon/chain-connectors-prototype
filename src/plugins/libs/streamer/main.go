package streamer

import (
	"context"
	"errors"
	"log"
	"math/big"
	"sync"

	"github.com/chris-de-leon/chain-connectors-prototype/src/plugins/libs/cursor"
)

var ErrStreamerStopped = errors.New("streamer has been stopped")

type Streamer struct {
	logger    *log.Logger
	signal    *sync.Cond
	cursor    cursor.Cursor
	isStopped bool
}

func New(cursor cursor.Cursor, logger *log.Logger) *Streamer {
	return &Streamer{
		signal:    sync.NewCond(&sync.Mutex{}),
		logger:    logger,
		cursor:    cursor,
		isStopped: false,
	}
}

func (streamer *Streamer) Subscribe(ctx context.Context) error {
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

	streamer.logger.Printf("Waiting for new data...")
	return streamer.cursor.Subscribe(ctx, func(cursor *big.Int) {
		streamer.logger.Printf("Received new cursor: %s", cursor.String())
		streamer.signal.Broadcast()
	})
}

func (streamer *Streamer) WaitForNextCursor(ctx context.Context) error {
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

func (streamer *Streamer) GetNextCursor(ctx context.Context, curr *big.Int) (*big.Int, error) {
	if streamer.isStopped {
		return nil, ErrStreamerStopped
	}

	latestCursor, err := streamer.cursor.GetLatestValue(ctx)
	if err != nil {
		return nil, err
	}

	if curr == nil || curr.Cmp(latestCursor) == -1 {
		return latestCursor, nil
	}

	if err := streamer.WaitForNextCursor(ctx); err != nil {
		return nil, err
	} else {
		return new(big.Int).Add(latestCursor, new(big.Int).SetUint64(1)), nil
	}
}
