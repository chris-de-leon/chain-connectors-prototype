package flow_testutils

import (
	"github.com/onflow/flow-go-sdk"
	"github.com/onflow/flow-go-sdk/access/grpc"
	"github.com/onflow/flow-go-sdk/crypto"
)

var (
	EMULATOR_SERVICE_ACCT_PRIV_KEY = "aff3a277caf2bdd6582c156ae7b07dbca537da7833309de88e56987faa2c0f1b"
	EMULATOR_SERVICE_ACCT_ADDR_HEX = "f8d6e0586b0a20c7"
	EMULATOR_SERVICE_ACCT_SIGN_ALG = crypto.ECDSA_P256
	EMULATOR_SERVICE_ACCT_HASH_ALG = crypto.SHA3_256
)

type (
	Account struct {
		PrivateKey crypto.PrivateKey
		Signer     crypto.InMemorySigner
		Address    flow.Address
	}
)

func NewEmulatorAccount() (*Account, error) {
	privateKey, err := crypto.DecodePrivateKeyHex(EMULATOR_SERVICE_ACCT_SIGN_ALG, EMULATOR_SERVICE_ACCT_PRIV_KEY)
	if err != nil {
		return nil, err
	}

	signer, err := crypto.NewInMemorySigner(privateKey, EMULATOR_SERVICE_ACCT_HASH_ALG)
	if err != nil {
		return nil, err
	}

	address := flow.HexToAddress(EMULATOR_SERVICE_ACCT_ADDR_HEX)
	return &Account{
		PrivateKey: privateKey,
		Signer:     signer,
		Address:    address,
	}, nil
}

func (acct *Account) SetBackend(backend *grpc.Client) *AugmentedAccount {
	return &AugmentedAccount{
		acct,
		backend,
	}
}
