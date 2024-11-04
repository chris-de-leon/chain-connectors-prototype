package geth

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient/simulated"
)

func InitBackend(acct *Account) (*simulated.Backend, error) {
	balance, ok := new(big.Int).SetString("10000000000000000000", 10) // 10 eth in wei
	if !ok {
		return nil, errors.New("failed to set genesis account balance")
	}

	return simulated.NewBackend(types.GenesisAlloc{
		acct.Addr(): types.Account{
			Balance: balance,
		},
	}), nil
}
