package staking

import (
	"context"
)

type StakingRepository interface {
	LoadAllPools(ctx context.Context) (map[string]*PoolAggregate, error)
	Save(ctx context.Context, blockNumber uint64, pool ...*PoolAggregate) error
}
