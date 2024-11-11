package eth_testutils

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
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

func (acct *AugmentedAccount) SignSendTx(ctx context.Context, tx *types.Transaction) (common.Hash, error) {
	signedTx, err := types.SignTx(tx, types.NewLondonSigner(tx.ChainId()), acct.privateKey)
	if err != nil {
		return signedTx.Hash(), err
	} else {
		return signedTx.Hash(), acct.backend.Client().SendTransaction(ctx, signedTx)
	}
}
