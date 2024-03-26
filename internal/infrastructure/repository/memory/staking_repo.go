package memory

import (
	"context"

	"github.com/IErcOrg/IERC_Indexer/internal/domain/staking"
	rctx "github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/context"
)

type stakingMemoryRepo struct {
	repo  staking.StakingRepository
	pools map[string]*staking.PoolAggregate
}

func (s *stakingMemoryRepo) LoadAllPools(ctx context.Context) (map[string]*staking.PoolAggregate, error) {
	var pools = make(map[string]*staking.PoolAggregate)
	for key, aggregate := range s.pools {
		pools[key] = aggregate
	}

	return pools, nil
}

func (s *stakingMemoryRepo) Save(ctx context.Context, blockNumber uint64, pools ...*staking.PoolAggregate) error {
	updateKind := rctx.UpdateKindFromContext(ctx)
	switch updateKind {
	case rctx.UpdateCache:

		for _, root := range pools {
			if root.Owner == "" {
				continue
			}

			s.pools[root.PoolAddress] = root
		}
		return nil

	case rctx.UpdateDB:
		return s.repo.Save(ctx, blockNumber, pools...)

	default:
		return nil
	}
}

func NewStakingMemoryRepository(repo staking.StakingRepository) (staking.StakingRepository, error) {

	ctx := context.Background()
	roots, err := repo.LoadAllPools(ctx)
	if err != nil {
		return nil, err
	}

	if roots == nil {
		roots = make(map[string]*staking.PoolAggregate)
	}

	srv := &stakingMemoryRepo{
		repo:  repo,
		pools: roots,
	}

	return srv, nil
}
