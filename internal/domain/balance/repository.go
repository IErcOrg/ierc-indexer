package balance

import (
	"context"
)

type BalanceRepository interface {
	Save(ctx context.Context, entities ...*Balance) error
	Load(ctx context.Context, key BalanceKey) (*Balance, error)
}
