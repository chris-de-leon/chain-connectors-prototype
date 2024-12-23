package flow_testutils

import (
	"context"
	"crypto/rand"
	"errors"
	"time"

	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/access/grpc"
	"github.com/onflow/flow-go-sdk/crypto"
	"github.com/onflow/flow-go-sdk/templates"
)

type AugmentedAccount struct {
	*Account
	Backend *grpc.Client
}

type TransactionResult[T any] struct {
	Receipt *flow.TransactionResult
	Result  T
}

func (acct *AugmentedAccount) SignTx(tx *flow.Transaction, keyIndex uint32) (*flow.Transaction, error) {
	return tx, tx.SignEnvelope(acct.Address, keyIndex, acct.Signer)
}

func (acct *AugmentedAccount) SendTx(ctx context.Context, tx *flow.Transaction) (*flow.Transaction, error) {
	return tx, acct.Backend.SendTransaction(ctx, *tx)
}

func (acct *AugmentedAccount) WaitTx(ctx context.Context, tx *flow.Transaction) (*flow.Transaction, error) {
	status := flow.TransactionStatusUnknown

	for status != flow.TransactionStatusSealed {
		result, err := acct.Backend.GetTransactionResult(ctx, tx.ID())
		if err != nil {
			return nil, err
		} else {
			status = result.Status
		}
		time.Sleep(time.Second)
	}

	return tx, nil
}

func (acct *AugmentedAccount) SignAndSendTx(ctx context.Context, tx *flow.Transaction, keyIndex uint32) (*flow.Transaction, error) {
	if tx, err := acct.SignTx(tx, keyIndex); err != nil {
		return nil, err
	} else {
		return acct.SendTx(ctx, tx)
	}
}

func (acct *AugmentedAccount) CreateAccount(ctx context.Context) (*TransactionResult[*Account], error) {
	signAlgo := crypto.ECDSA_P256
	hashAlgo := crypto.SHA3_256

	seed := make([]byte, crypto.MinSeedLength)
	if _, err := rand.Read(seed); err != nil {
		return nil, err
	}

	privateKey, err := crypto.GeneratePrivateKey(signAlgo, seed)
	if err != nil {
		return nil, err
	}

	signer, err := crypto.NewInMemorySigner(privateKey, hashAlgo)
	if err != nil {
		return nil, err
	}

	accountKey := flow.NewAccountKey().
		FromPrivateKey(privateKey).
		SetHashAlgo(crypto.SHA3_256).
		SetWeight(flow.AccountKeyWeightThreshold)

	tx, err := templates.CreateAccount([]*flow.AccountKey{accountKey}, nil, acct.Address)
	if err != nil {
		return nil, err
	}

	block, err := acct.Backend.GetLatestBlock(ctx, true)
	if err != nil {
		return nil, err
	}

	info, err := acct.Backend.GetAccount(ctx, acct.Address)
	if err != nil {
		return nil, err
	}

	var key *flow.AccountKey
	if len(info.Keys) == 0 {
		return nil, errors.New("no account keys available")
	} else {
		key = info.Keys[0]
	}

	tx = tx.SetProposalKey(acct.Address, key.Index, key.SequenceNumber).
		SetReferenceBlockID(block.ID).
		SetPayer(acct.Address)

	if tx, err = acct.SignAndSendTx(ctx, tx, key.Index); err != nil {
		return nil, err
	}

	if tx, err = acct.WaitTx(ctx, tx); err != nil {
		return nil, err
	}

	receipt, err := acct.Backend.GetTransactionResult(ctx, tx.ID())
	if err != nil {
		return nil, err
	}

	for _, event := range receipt.Events {
		if event.Type == flow.EventAccountCreated {
			address := flow.AccountCreatedEvent(event).Address()
			return &TransactionResult[*Account]{
				Receipt: receipt,
				Result: &Account{
					PrivateKey: privateKey,
					Signer:     signer,
					Address:    address,
				},
			}, nil
		}
	}

	return nil, errors.New("failed to extract account address from transaction receipt")
}
