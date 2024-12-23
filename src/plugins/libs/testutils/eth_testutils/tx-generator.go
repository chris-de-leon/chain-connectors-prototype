package eth_testutils

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

type TransactionGenerator struct {
	account *AugmentedAccount
	logger  *log.Logger
}

func NewTransactionGenerator(acct *AugmentedAccount, logger *log.Logger) *TransactionGenerator {
	return &TransactionGenerator{account: acct, logger: logger}
}

func NewTransactionGeneratorLogger() *log.Logger {
	return log.New(os.Stdout, fmt.Sprintf("[%s] ", "transaction-generator"), log.LstdFlags)
}

func (generator *TransactionGenerator) Start(ctx context.Context, interval time.Duration, count int) error {
	timer := time.NewTimer(interval)
	defer timer.Stop()
	for {
		// Suppose instead of a timer we used a 2-second ticker but it takes 3+ seconds
		// to perform processing. If the ticker activates while data is being processed,
		// then we'll immediately process the data again, which is not intended. Instead,
		// we want to fully wait another 2 seconds *from the time we finished processing
		// the last round* before trying to process the data again. With that in mind a
		// timer would be more appropriate here.
		timer.Reset(interval)
		select {
		case <-ctx.Done():
			return nil
		case _, ok := <-timer.C:
			if !ok {
				return nil
			}
			for i := 0; i < count; i++ {
				if tx, err := generator.account.TransferTokens(ctx, generator.account.Address, 1); err != nil {
					return err
				} else {
					generator.logger.Printf("New transaction: %s", tx.Hash().String())
				}
			}
		}
	}
}
