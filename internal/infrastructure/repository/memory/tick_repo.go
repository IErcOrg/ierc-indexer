package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/tick"
	rctx "github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/context"
	"github.com/allegro/bigcache"
)

const (
	TickCacheKeyPrefix = "tick####"
)

type message struct {
	Protocol protocol.Protocol `json:"protocol,omitempty"`
	Data     []byte            `json:"data,omitempty"`
}

type tickMemoryRepo struct {
	db    tick.TickRepository
	cache *bigcache.BigCache
	mutex sync.Mutex
}

func (repo *tickMemoryRepo) Save(ctx context.Context, entities ...tick.Tick) error {

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

func (repo *tickMemoryRepo) Load(ctx context.Context, tickName string) (tick.Tick, error) {

	entity, err := repo.getCache(tickName)
	if err == nil {
		return entity, nil
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	entity, err = repo.getCache(tickName)
	if err == nil {
		return entity, nil
	}

	entity, err = repo.db.Load(ctx, tickName)
	if err != nil {
		return nil, err
	}

	if entity != nil {
		repo.setCache(entity)
	}

	return entity, nil
}

func (repo *tickMemoryRepo) updateCache(entities ...tick.Tick) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()
	for _, entity := range entities {

		if entity.GetID() == 0 {
			continue
		}

		repo.setCache(entity)
	}

	return nil
}

func (repo *tickMemoryRepo) setCache(entity tick.Tick) {
	bytes, err := entity.Marshal()
	if err != nil {
		return
	}

	msg := message{
		Protocol: entity.GetProtocol(),
		Data:     bytes,
	}
	msgBytes, _ := json.Marshal(msg)
	key := fmt.Sprintf("%s_%s", TickCacheKeyPrefix, entity.GetName())
	_ = repo.cache.Set(key, msgBytes)
}

func (repo *tickMemoryRepo) getCache(key string) (tick.Tick, error) {
	bytes, err := repo.cache.Get(fmt.Sprintf("%s_%s", TickCacheKeyPrefix, key))
	if err != nil {
		return nil, err
	}

	var m message
	if err = json.Unmarshal(bytes, &m); err != nil {
		return nil, err
	}

	switch m.Protocol {
	case protocol.ProtocolIERCPoW:
		entity := new(tick.IERCPoWTick)
		return entity, entity.Unmarshal(m.Data)

	default:
		entity := new(tick.IERC20Tick)
		return entity, entity.Unmarshal(m.Data)
	}
}

func NewTickMemoryRepository(db tick.TickRepository, cache *bigcache.BigCache) tick.TickRepository {
	return &tickMemoryRepo{
		db:    db,
		cache: cache,
		mutex: sync.Mutex{},
	}
}
