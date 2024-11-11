package solana_testutils

import (
	"github.com/gagliardetto/solana-go"
)

type Account struct {
	PrivateKey solana.PrivateKey
}

func NewAccount() (*Account, error) {
	privateKey, err := solana.NewRandomPrivateKey()
	if err != nil {
		return nil, err
	}

	return &Account{
		PrivateKey: privateKey,
	}, nil
}

func (acct *Account) SetBackend(backend *Backend) *AugmentedAccount {
	return &AugmentedAccount{
		acct,
		backend,
	}
}
