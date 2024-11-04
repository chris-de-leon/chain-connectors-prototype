package geth

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
)

type AugmentedAccount struct {
	*Account
	backend *simulated.Backend
}

func (acct *AugmentedAccount) Backend() *simulated.Backend {
	return acct.backend
}

func (acct *AugmentedAccount) SignSendTx(ctx context.Context, tx *types.Transaction) error {
	signedTx, err := types.SignTx(tx, types.NewLondonSigner(tx.ChainId()), acct.privateKey)
	if err != nil {
		return err
	} else {
		return acct.backend.Client().SendTransaction(ctx, signedTx)
	}
}
