package geth

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
)

type BlockGenerator struct {
	account *AugmentedAccount
	logger  *log.Logger
}

func NewBlockGenerator(acct *AugmentedAccount, logger *log.Logger) *BlockGenerator {
	return &BlockGenerator{account: acct, logger: logger}
}

func NewBlockGeneratorLogger() *log.Logger {
	return log.New(os.Stdout, fmt.Sprintf("[%s] ", "block-generator"), log.LstdFlags)
}

func (generator *BlockGenerator) Start(ctx context.Context, interval time.Duration, count int) error {
	timerDuration := time.Duration(interval) * time.Millisecond
	timer := time.NewTimer(timerDuration)
	defer timer.Stop()
	for {
		// Suppose instead of a timer we used a 2-second ticker but it takes 3+ seconds
		// to perform processing. If the ticker activates while data is being processed,
		// then we'll immediately process the data again, which is not intended. Instead,
		// we want to fully wait another 2 seconds *from the time we finished processing
		// the last round* before trying to process the data again. With that in mind a
		// timer would be more appropriate here.
		timer.Reset(timerDuration)
		select {
		case <-ctx.Done():
			return nil
		case _, ok := <-timer.C:
			if !ok {
				return nil
			}
			for i := 0; i < count; i++ {
				if err := generator.sendDummyTransaction(ctx); err != nil {
					return err
				}
			}
		}
	}
}

func (generator *BlockGenerator) sendDummyTransaction(ctx context.Context) error {
	oneETHER := new(big.Int).Mul(big.NewInt(1), big.NewInt(params.Ether))
	addr := generator.account.Addr()

	nonce, err := generator.account.Backend().Client().PendingNonceAt(ctx, addr)
	if err != nil {
		return err
	}

	gasTipCap, err := generator.account.Backend().Client().SuggestGasTipCap(ctx)
	if err != nil {
		return err
	}

	gasFeeCap, err := generator.account.Backend().Client().SuggestGasPrice(ctx)
	if err != nil {
		return err
	}

	chainID, err := generator.account.backend.Client().ChainID(ctx)
	if err != nil {
		return err
	}

	if err = generator.account.SignSendTx(
		ctx,
		types.NewTx(
			&types.DynamicFeeTx{
				ChainID:   chainID,
				Nonce:     nonce,
				GasTipCap: gasTipCap,
				GasFeeCap: gasFeeCap,
				Gas:       uint64(21000),
				To:        &addr,
				Value:     oneETHER,
				Data:      nil,
			},
		),
	); err != nil {
		return err
	} else {
		defer generator.logger.Printf("New block: %s", generator.account.Backend().Commit())
	}

	return nil
}
