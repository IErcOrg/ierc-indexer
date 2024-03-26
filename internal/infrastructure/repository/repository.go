package repository

import (
	"github.com/IErcOrg/IERC_Indexer/internal/domain/balance"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol/parser"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/staking"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/tick"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/memory"
	mysqlimpl "github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/mysql"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/network/ethereum"
	"github.com/allegro/bigcache"
	"github.com/google/wire"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(
	NewDB,
	NewCache,
	NewData,
	NewTransactionRepository,
	NewProtocolParser,
	NewEthereumFetcher,
	NewBlockRepository,
	NewTickRepository,
	NewBalanceRepository,
	NewEventRepository,
	NewStakingRepository,
)

var (
	NewProtocolParser  = parser.NewParser
	NewEthereumFetcher = ethereum.NewEthereumFetcher
	NewBlockRepository = mysqlimpl.NewBlockRepo
	NewEventRepository = mysqlimpl.NewEventRepository
)

func NewTickRepository(db *gorm.DB, cache *bigcache.BigCache) tick.TickRepository {
	return memory.NewTickMemoryRepository(mysqlimpl.NewTickRepo(db), cache)
}

func NewBalanceRepository(db *gorm.DB, cache *bigcache.BigCache) balance.BalanceRepository {
	return memory.NewBalanceMemoryRepository(mysqlimpl.NewBalanceRepo(db), cache)
}

func NewStakingRepository(db *gorm.DB) (staking.StakingRepository, error) {
	return memory.NewStakingMemoryRepository(mysqlimpl.NewStakingRepository(db))
}
