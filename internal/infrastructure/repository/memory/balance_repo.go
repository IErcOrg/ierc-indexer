package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/IErcOrg/IERC_Indexer/internal/domain/balance"
	rctx "github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/context"
	"github.com/allegro/bigcache"
)

const (
	BalanceCacheKeyPrefix = "balance####"
)

type balanceMemoryRepo struct {
	db    balance.BalanceRepository
	cache *bigcache.BigCache
	mutex sync.Mutex
}

func (repo *balanceMemoryRepo) Save(ctx context.Context, entities ...*balance.Balance) error {
	updateKind := rctx.UpdateKindFromContext(ctx)
	switch updateKind {
	case rctx.UpdateCache:
		return repo.updateCache(entities...)

	case rctx.UpdateDB:
		return repo.db.Save(ctx, entities...)
	default:
		return nil
	}
}

func (repo *balanceMemoryRepo) Load(ctx context.Context, key balance.BalanceKey) (*balance.Balance, error) {

	entity, err := repo.getCache(key.String())
	if err == nil {
		return entity, nil
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	entity, err = repo.getCache(key.String())
	if err == nil {
		return entity, nil
	}

	entity, err = repo.db.Load(ctx, key)
	if err != nil {
		return nil, err // database error
	}

	if entity != nil {
		repo.setCache(entity)
	}

	return entity, nil
}

func (repo *balanceMemoryRepo) updateCache(entities ...*balance.Balance) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	for _, entity := range entities {

		if entity.ID == 0 {
			continue
		}

		repo.setCache(entity)
	}

	return nil
}

func (repo *balanceMemoryRepo) setCache(entity *balance.Balance) {
	bytes, err := entity.Marshal()
	if err != nil {
		return
	}

	key := fmt.Sprintf("%s_%s", BalanceCacheKeyPrefix, entity.Key())
	_ = repo.cache.Set(key, bytes)
}

func (repo *balanceMemoryRepo) getCache(key string) (*balance.Balance, error) {
	bytes, err := repo.cache.Get(fmt.Sprintf("%s_%s", BalanceCacheKeyPrefix, key))
	if err != nil {
		return nil, err
	}

	entity := new(balance.Balance)
	return entity, entity.Unmarshal(bytes)
}

func NewBalanceMemoryRepository(db balance.BalanceRepository, cache *bigcache.BigCache) balance.BalanceRepository {
	return &balanceMemoryRepo{
		db:    db,
		cache: cache,
		mutex: sync.Mutex{},
	}
}
