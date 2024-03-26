package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type StakingPool struct {
	ID               int64     `gorm:"<-:create;column:id;primaryKey;autoIncrement"`
	Pool             string    `gorm:"<-:create;column:pool;type:varchar(42);uniqueIndex:uni_pool,priority:1"`
	PoolID           uint64    `gorm:"<-:create;column:pool_id;type:bigint;uniqueIndex:uni_pool,priority:2"`
	Name             string    `gorm:"column:name;type:varchar(64)"`
	Owner            string    `gorm:"column:owner;type:varchar(42);index:id_owner;not null"`
	Data             []byte    `gorm:"column:data;type:json"`
	LastUpdatedBlock uint64    `gorm:"column:last_updated_block;type:bigint"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (t *StakingPool) TableName() string {
	return "staking_pools"
}

type StakingPosition struct {
	ID               int64           `gorm:"<-:create;column:id;primaryKey;autoIncrement"`
	Pool             string          `gorm:"<-:create;column:pool;type:varchar(42);uniqueIndex:uni_pool_staker,priority:1;not null"`
	PoolID           uint64          `gorm:"<-:create;column:pool_id;type:bigint;uniqueIndex:uni_pool_staker,priority:2"`
	Staker           string          `gorm:"<-:create;column:staker;type:varchar(42);uniqueIndex:uni_pool_staker,priority:3;not null"`
	AccRewards       decimal.Decimal `gorm:"column:acc_rewards;type:decimal(50,18);not null;default:0.000000000000000000"`
	Debt             decimal.Decimal `gorm:"column:debt;type:decimal(50,18);not null;default:0.000000000000000000"`
	RewardsPerBlock  decimal.Decimal `gorm:"column:rewards_per_block;type:decimal(50,18);not null;default:0.000000000000000000"`
	LastRewardBlock  uint64          `gorm:"column:last_reward_block;type:bigint"`
	LastUpdatedBlock uint64          `gorm:"column:last_updated_block;type:bigint"`
	Amounts          []byte          `gorm:"column:staker_amounts;type:json"`
	CreatedAt        time.Time       `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt        time.Time       `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (t *StakingPosition) TableName() string {
	return "staking_positions"
}

type StakingBalance struct {
	ID          int64           `gorm:"<-:create;column:id;primaryKey;autoIncrement"`
	Staker      string          `gorm:"<-:create;column:staker;type:varchar(42);uniqueIndex:uni_staker_pool_tick,priority:1;not null"`
	Pool        string          `gorm:"<-:create;column:pool;type:varchar(42);uniqueIndex:uni_staker_pool_tick,priority:2;index:idx_pool;not null"`
	PoolID      uint64          `gorm:"<-:create;column:pool_id;type:bigint;uniqueIndex:uni_staker_pool_tick,priority:3;index:idx_pool_id"`
	Tick        string          `gorm:"<-:create;column:tick;type:varchar(64);uniqueIndex:uni_staker_pool_tick,priority:4;not null;default:''"`
	Amount      decimal.Decimal `gorm:"column:amount;type:decimal(50,18);not null;default:0.000000000000000000"`
	BlockNumber uint64          `gorm:"column:block_number;type:bigint"`
	CreatedAt   time.Time       `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt   time.Time       `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (t *StakingBalance) TableName() string {
	return "staking_balances"
}
