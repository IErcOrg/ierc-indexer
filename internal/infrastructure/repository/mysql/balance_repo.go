package mysqlimpl

import (
	"context"
	"errors"

	"github.com/IErcOrg/IERC_Indexer/internal/domain/balance"
	rctx "github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/context"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/mysql/acl"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/mysql/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type balanceMySQLRepo struct {
	db *gorm.DB
}

func NewBalanceRepo(db *gorm.DB) balance.BalanceRepository {
	return &balanceMySQLRepo{db: db}
}

func (repo *balanceMySQLRepo) Load(ctx context.Context, key balance.BalanceKey) (*balance.Balance, error) {

	var m models.IERC20Balance
	err := repo.db.WithContext(ctx).Where("address = ? and tick = ?", key.Address, key.Tick).Take(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return acl.ConvertBalanceModelToEntity(&m), nil
}

func (repo *balanceMySQLRepo) Save(ctx context.Context, entities ...*balance.Balance) error {
	if len(entities) == 0 {
		return nil
	}

	db := rctx.TransactionDBFromContext(ctx)
	if db == nil {
		panic("missing db instance")
	}

	var ms []*models.IERC20Balance
	for _, entity := range entities {
		ms = append(ms, acl.ConvertBalanceEntityToModel(entity))
	}

	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: `id`}},
		DoUpdates: clause.AssignmentColumns([]string{
			`available`,
			`freeze`,
			`minted`,
			`last_updated_block`,
			`updated_at`,
		}),
	}).CreateInBatches(ms, 1000).Error
}
