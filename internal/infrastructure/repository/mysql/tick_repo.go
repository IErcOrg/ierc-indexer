package mysqlimpl

import (
	"context"
	"errors"

	domain "github.com/IErcOrg/IERC_Indexer/internal/domain/tick"
	rctx "github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/context"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/mysql/acl"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/mysql/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type tickRepo struct {
	db *gorm.DB
}

func (repo *tickRepo) Load(ctx context.Context, tick string) (domain.Tick, error) {
	// query
	var m models.IERCTick
	if err := repo.db.WithContext(ctx).Where("tick = ?", tick).Take(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return acl.ConvertTickModelToEntity(&m)
}

func (repo *tickRepo) Save(ctx context.Context, entities ...domain.Tick) error {

	if len(entities) == 0 {
		return nil
	}

	db := rctx.TransactionDBFromContext(ctx)
	if db == nil {
		panic("missing db instance")
	}

	var ms []*models.IERCTick
	for _, entity := range entities {
		ms = append(ms, acl.ConvertTickEntityToModel(entity))
	}

	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: `id`}},
		DoUpdates: clause.AssignmentColumns([]string{
			`max_supply`,
			`supply`,
			`detail`,
			`last_updated_block`,
			`updated_at`,
		}),
	}).CreateInBatches(ms, 1000).Error
}

func NewTickRepo(db *gorm.DB) domain.TickRepository {
	return &tickRepo{db: db}
}
