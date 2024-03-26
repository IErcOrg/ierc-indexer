package mysqlimpl

import (
	"context"

	"github.com/IErcOrg/IERC_Indexer/internal/domain/staking"
	rctx "github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/context"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/mysql/acl"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/mysql/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type stakingRepo struct {
	db *gorm.DB
}

func (repo *stakingRepo) LoadAllPools(ctx context.Context) (map[string]*staking.PoolAggregate, error) {

	var pools []string
	err := repo.db.WithContext(ctx).
		Table((&models.StakingPool{}).TableName()).
		Select("distinct pool").Find(&pools).Error
	if err != nil {
		return nil, err
	}

	var poolRoots = make(map[string]*staking.PoolAggregate)
	for _, pool := range pools {
		poolRoot, err := repo.QueryPoolAggregate(ctx, pool)
		if err != nil {
			return nil, err
		}

		if poolRoots == nil {
			continue
		}

		if err := repo.LoadAllPositionsByPool(ctx, poolRoot); err != nil {
			return nil, err
		}

		poolRoots[poolRoot.PoolAddress] = poolRoot
	}

	return poolRoots, nil
}

func (repo *stakingRepo) QueryPoolAggregate(ctx context.Context, pool string) (*staking.PoolAggregate, error) {
	var ms []*models.StakingPool
	if err := repo.db.WithContext(ctx).Where("pool = ?", pool).Find(&ms).Error; err != nil {
		return nil, err
	}

	if len(ms) == 0 {
		return nil, nil
	}

	m := ms[0]
	var aggregate = staking.NewPoolAggregate(m.Pool, m.Owner)
	for _, stakingPool := range ms {
		entity, err := acl.ConvertPoolModelToEntity(stakingPool)
		if err != nil {
			continue
		}

		aggregate.InitPool(entity)
	}

	return aggregate, nil
}

func (repo *stakingRepo) LoadAllPositionsByPool(ctx context.Context, pool *staking.PoolAggregate) error {

	var ms []*models.StakingPosition

	return repo.db.WithContext(ctx).Where("pool = ?", pool.PoolAddress).
		FindInBatches(&ms, 1000, func(tx *gorm.DB, batch int) error {
			for _, m := range ms {
				pool.InitPosition(acl.ConvertPositionModelToEntity(m))
			}

			return nil
		}).Error
}

func (repo *stakingRepo) Save(ctx context.Context, blockNumber uint64, roots ...*staking.PoolAggregate) error {
	if len(roots) == 0 {
		return nil
	}

	db := rctx.TransactionDBFromContext(ctx)
	if db == nil {
		panic("missing db instance")
	}

	var (
		pools     []*models.StakingPool
		positions []*models.StakingPosition
		balances  []*models.StakingBalance
	)

	for _, root := range roots {

		var entities = root.GetStakingPools()
		for _, pool := range entities {
			if pool.LastUpdatedBlock < blockNumber {
				continue
			}

			pools = append(pools, acl.ConvertPoolEntityToModel(pool))
		}

		var stakingPositions = root.GetStakingPositions()
		for _, position := range stakingPositions {
			if position.LastUpdatedBlock < blockNumber {
				continue
			}

			positions = append(positions, acl.ConvertPositionEntityToModel(position))

			balancesMap := acl.ConvertPositionEntityToBalanceModel(position)
			for _, balance := range balancesMap {
				balances = append(balances, balance)
			}
		}
	}

	if len(pools) != 0 {
		err := db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: `pool`}, {Name: `pool_id`}},
			DoUpdates: clause.AssignmentColumns([]string{
				`name`,
				`owner`,
				`data`,
				`last_updated_block`,
				`updated_at`,
			}),
		}).CreateInBatches(pools, 1000).Error
		if err != nil {
			return err
		}
	}

	if len(balances) != 0 {
		err := db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: `staker`}, {Name: `pool`}, {Name: `pool_id`}, {Name: `tick`}},
			DoUpdates: clause.AssignmentColumns([]string{`amount`, `block_number`, `updated_at`}),
		}).CreateInBatches(balances, 1000).Error
		if err != nil {
			return err
		}
	}

	if len(positions) != 0 {
		err := db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: `pool`}, {Name: `pool_id`}, {Name: `staker`}},
			DoUpdates: clause.AssignmentColumns([]string{
				`acc_rewards`,
				`debt`,
				`rewards_per_block`,
				`last_reward_block`,
				`last_updated_block`,
				`staker_amounts`,
				`updated_at`,
			}),
		}).CreateInBatches(positions, 1000).Error
		if err != nil {
			return err
		}
	}

	return nil
}

func NewStakingRepository(db *gorm.DB) staking.StakingRepository {
	return &stakingRepo{db: db}
}
