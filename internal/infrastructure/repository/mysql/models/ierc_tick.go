package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type IERCTick struct {
	ID               int64           `gorm:"<-:create;column:id;primaryKey;autoIncrement"`
	Protocol         string          `gorm:"<-:create;column:protocol;type:varchar(20);not null;default:''"`
	Tick             string          `gorm:"<-:create;column:tick;type:varchar(64);uniqueIndex:idx_tick;not null;default:''"`
	Decimals         int64           `gorm:"<-:create;column:decimals;type:int;not null;default:0"`
	Creator          string          `gorm:"<-:create;column:creator;type:varchar(64);not null;default:''"`
	MaxSupply        decimal.Decimal `gorm:"<-:create;column:max_supply;type:decimal(50,18);not null;default:0.000000000000000000"`
	Supply           decimal.Decimal `gorm:"column:supply;type:decimal(50,18);not null;default:0.000000000000000000"`
	Detail           []byte          `gorm:"column:detail;type:json"`
	LastUpdatedBlock uint64          `gorm:"column:last_updated_block;type:bigint"`
	CreatedAt        time.Time       `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt        time.Time       `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (t *IERCTick) TableName() string {
	return "ierc_ticks"
}
