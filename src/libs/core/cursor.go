package core

import (
	"context"
	"math/big"
)

type Cursor interface {
	GetLatestValue(ctx context.Context) (*big.Int, error)
	Subscribe(ctx context.Context, cb func(cursor *big.Int)) error
}
