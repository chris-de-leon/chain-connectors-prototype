package eth_testutils

import (
	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
)

type Account struct {
	PrivateKey *ecdsa.PrivateKey
	Address    common.Address
}

func NewAccount() (*Account, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	return &Account{
		PrivateKey: privateKey,
		Address:    crypto.PubkeyToAddress(privateKey.PublicKey),
	}, nil
}

func (acct *Account) SetBackend(backend *simulated.Backend) *AugmentedAccount {
	return &AugmentedAccount{
		acct,
		backend,
	}
}
