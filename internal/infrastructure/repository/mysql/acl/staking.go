package acl

import (
	"encoding/json"
	"time"

	"github.com/IErcOrg/IERC_Indexer/internal/domain/staking"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/mysql/models"
)

func ConvertPoolEntityToModel(pool *staking.StakingPool) *models.StakingPool {
	data, _ := json.Marshal(pool)

	return &models.StakingPool{
		ID:               0,
		Pool:             pool.Pool,
		PoolID:           pool.PoolSubID,
		Name:             pool.Detail.Name,
		Owner:            pool.Detail.Owner,
		Data:             data,
		LastUpdatedBlock: pool.LastUpdatedBlock,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func ConvertPoolModelToEntity(m *models.StakingPool) (*staking.StakingPool, error) {
	var pool staking.StakingPool
	return &pool, json.Unmarshal(m.Data, &pool)
}

func ConvertPositionEntityToModel(position *staking.StakingPosition) *models.StakingPosition {
	bytes, _ := json.Marshal(position.TickDetails)

	return &models.StakingPosition{
		ID:               0,
		Pool:             position.PoolAddress,
		PoolID:           position.PoolSubID,
		Staker:           position.Staker,
		AccRewards:       position.AccReward,
		Debt:             position.Debt,
		RewardsPerBlock:  position.RewardsPerBlock,
		LastRewardBlock:  position.LastRewardBlock,
		LastUpdatedBlock: position.LastUpdatedBlock,
		Amounts:          bytes,
		CreatedAt:        position.CreatedAt,
		UpdatedAt:        position.UpdatedAt,
	}
}

func ConvertPositionModelToEntity(m *models.StakingPosition) *staking.StakingPosition {
	var amounts = make(map[string]*staking.PositionTickDetail)
	_ = json.Unmarshal(m.Amounts, &amounts)

	return &staking.StakingPosition{
		PoolAddress:      m.Pool,
		PoolSubID:        m.PoolID,
		Staker:           m.Staker,
		TickDetails:      amounts,
		RewardsPerBlock:  m.RewardsPerBlock,
		Debt:             m.Debt,
		AccReward:        m.AccRewards,
		LastRewardBlock:  m.LastRewardBlock,
		LastUpdatedBlock: m.LastUpdatedBlock,
		CreatedAt:        m.CreatedAt,
		UpdatedAt:        m.UpdatedAt,
	}
}

func ConvertPositionEntityToBalanceModel(position *staking.StakingPosition) map[string]*models.StakingBalance {
	var balancesMap = make(map[string]*models.StakingBalance)

	for _, detail := range position.TickDetails {
		balancesMap[detail.Tick] = &models.StakingBalance{
			ID:          0,
			Staker:      position.Staker,
			Pool:        position.PoolAddress,
			PoolID:      position.PoolSubID,
			Tick:        detail.Tick,
			Amount:      detail.Amount,
			BlockNumber: position.LastUpdatedBlock,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
	}

	return balancesMap
}
