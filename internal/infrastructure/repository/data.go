package repository

import (
	"context"
	"time"

	"github.com/IErcOrg/IERC_Indexer/internal/conf"
	"github.com/IErcOrg/IERC_Indexer/internal/domain"
	rctx "github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/context"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/mysql/models"
	"github.com/allegro/bigcache"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewCache() (*bigcache.BigCache, func(), error) {
	cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(time.Minute * 10))
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		_ = cache.Close()
	}

	return cache, cleanup, nil
}

func NewDB(c *conf.Config, l log.Logger) (*gorm.DB, func(), error) {

	data := c.Bootstrap.GetData()

	helper := log.NewHelper(l)

	helper.Info("initial mysql")
	inner, err := gorm.Open(mysql.Open(data.Database.Source), &gorm.Config{})
	if err != nil {
		return nil, nil, err
	}

	inner.Logger = inner.Logger.LogMode(logger.LogLevel(data.Database.LogLevel))

	db, err := inner.DB()
	if err != nil {
		return nil, nil, err
	}

	if data.Database.MaxIdleConns != 0 {
		db.SetMaxIdleConns(int(data.Database.MaxIdleConns))
	} else {
		db.SetMaxIdleConns(10)
	}

	if data.Database.MaxOpenConns != 0 {
		db.SetMaxOpenConns(int(data.Database.MaxOpenConns))
	} else {
		db.SetMaxOpenConns(100)
	}

	if data.Database.ConnMaxLifetime != nil {
		db.SetConnMaxLifetime(data.Database.ConnMaxLifetime.AsDuration())
	} else {
		db.SetConnMaxLifetime(time.Second * 300)
	}

	cleanup := func() {
		_ = db.Close()
	}

	err = inner.
		Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4 COLLATE=utf8mb4_bin").
		AutoMigrate(
			&models.Block{},
			&models.Transaction{},
			&models.Event{},
			&models.IERCTick{},
			&models.IERC20Balance{},
			&models.StakingPool{},
			&models.StakingPosition{},
			&models.StakingBalance{},
		)

	return inner, cleanup, err
}

type Data struct {
	db    *gorm.DB
	cache *bigcache.BigCache
}

func (d *Data) TransactionSave(ctx context.Context, fn func(ctx context.Context) error) error {
	return d.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctx := rctx.WithTransactionDB(ctx, tx)
		ctx = rctx.WithUpdateKind(ctx, rctx.UpdateDB)
		return fn(ctx)
	})
}

func (d *Data) UpdateCache(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(rctx.WithUpdateKind(ctx, rctx.UpdateCache))
}

func NewData(db *gorm.DB, cache *bigcache.BigCache) *Data {
	return &Data{db: db, cache: cache}
}

func NewTransactionRepository(data *Data) domain.TransactionRepository {
	return data
}
