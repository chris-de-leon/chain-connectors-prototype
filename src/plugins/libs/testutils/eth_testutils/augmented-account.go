package eth_testutils

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
	"github.com/ethereum/go-ethereum/params"
)

type AugmentedAccount struct {
	*Account
	Backend *simulated.Backend
}

func (acct *AugmentedAccount) SignTx(tx *types.Transaction) (*types.Transaction, error) {
	return types.SignTx(tx, types.NewLondonSigner(tx.ChainId()), acct.PrivateKey)
}

func (acct *AugmentedAccount) SendTx(ctx context.Context, tx *types.Transaction) (*types.Transaction, error) {
	if err := acct.Backend.Client().SendTransaction(ctx, tx); err != nil {
		return nil, err
	} else {
		return tx, err
	}
}

func (acct *AugmentedAccount) WaitTx(ctx context.Context, tx *types.Transaction) (*types.Transaction, error) {
	isPending := true

	for isPending {
		_, status, err := acct.Backend.Client().TransactionByHash(ctx, tx.Hash())
		if err != nil {
			return nil, err
		} else {
			isPending = status
		}
		time.Sleep(time.Second)
	}

	return tx, nil
}

func (acct *AugmentedAccount) SignAndSendTx(ctx context.Context, tx *types.Transaction) (*types.Transaction, error) {
	if tx, err := acct.SignTx(tx); err != nil {
		return nil, err
	} else {
		return acct.SendTx(ctx, tx)
	}
}

func (acct *AugmentedAccount) TransferTokens(ctx context.Context, recipient common.Address, ethers int64) (*types.Transaction, error) {
	amount := new(big.Int).Mul(big.NewInt(ethers), big.NewInt(params.Ether))

	nonce, err := acct.Backend.Client().PendingNonceAt(ctx, acct.Address)
	if err != nil {
		return nil, err
	}

	gasTipCap, err := acct.Backend.Client().SuggestGasTipCap(ctx)
	if err != nil {
		return nil, err
	}

	gasFeeCap, err := acct.Backend.Client().SuggestGasPrice(ctx)
	if err != nil {
		return nil, err
	}

	chainID, err := acct.Backend.Client().ChainID(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := acct.SignAndSendTx(ctx,
		types.NewTx(
			&types.DynamicFeeTx{
				ChainID:   chainID,
				Nonce:     nonce,
				GasTipCap: gasTipCap,
				GasFeeCap: gasFeeCap,
				Gas:       uint64(21000),
				To:        &recipient,
				Value:     amount,
				Data:      nil,
			},
		),
	)
	if err != nil {
		return nil, err
	} else {
		acct.Backend.Commit()
	}

	return acct.WaitTx(ctx, tx)
}
