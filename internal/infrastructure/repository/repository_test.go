package repository

import (
	"testing"

	"github.com/IErcOrg/IERC_Indexer/internal/domain"
	"github.com/stretchr/testify/suite"
)

const (
	MySqlDSN = "root:123456@(127.0.0.1:3306)/ierc_sepolia_indexer?charset=utf8mb4&parseTime=True&loc=Local"
)

func TestRepository(t *testing.T) {
	suite.Run(t, new(TestLRepositorySuite))
}

type TestLRepositorySuite struct {
	suite.Suite
	repo     domain.BlockRepository
	aggrRepo domain.EventRepository
}

func (s *TestLRepositorySuite) SetupSuite() {
	//
	//db, err := gorm.Open(mysql.Open(MySqlDSN), &gorm.Config{})
	//db.Logger = db.Logger.LogMode(logger.Info)
	//
	//s.Nil(err)
	//
	//s.repo = mysqlimpl.NewBlockRepo(db)
	//cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(time.Second * 10))
	//s.Nil(err)
	//s.cacheRepo = mysqlimpl.NewCacheRepository(cache)
	//s.aggrRepo = mysqlimpl.NewEventRepository(db, s.cacheRepo)
}

func (s *TestLRepositorySuite) TestQueryAddressHoldings() {

}

func (s *TestLRepositorySuite) TestSubscribe() {

}
