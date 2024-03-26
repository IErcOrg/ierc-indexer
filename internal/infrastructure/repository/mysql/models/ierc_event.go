package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Event struct {
	ID          int64           `gorm:"<-:create;column:id;primaryKey;autoIncrement"`
	BlockNumber uint64          `gorm:"<-:create;column:block_number;type:bigint;index:idx_block_number"`
	TxHash      string          `gorm:"<-:create;column:tx_hash;type:varchar(66);index:idx_hash;not null"`
	Operate     string          `gorm:"<-:create;column:operate;type:varchar(20);index:idx_operate;not null;default:''"`
	Tick        string          `gorm:"<-:create;column:tick;type:varchar(64);index:idx_tick;not null;default:''"`
	ETHFrom     string          `gorm:"<-:create;column:eth_from;type:varchar(42);index:idx_ierc_from;not null"`
	ETHTo       string          `gorm:"<-:create;column:eth_to;type:varchar(42);index:idx_ierc_to;not null"`
	IERCFrom    string          `gorm:"<-:create;column:ierc_from;type:varchar(42);index:idx_ierc_from;not null"`
	IERCTo      string          `gorm:"<-:create;column:ierc_to;type:varchar(42);index:idx_ierc_to;not null"`
	Amount      decimal.Decimal `gorm:"<-:create;column:amount;type:decimal(50,18);not null;default:0.000000000000000000"`
	Sign        string          `gorm:"<-:create;column:sign;type:varchar(256);index:idx_sign;not null;default:''"`
	EventKind   uint8           `gorm:"<-:create;column:event_kind;type:tinyint;not null;default:0"`
	Event       []byte          `gorm:"<-:create;column:event_data;type:json"`
	ErrCode     int32           `gorm:"<-:create;column:err_code;type:int;index:idx_err_code;not null;default:0"`
	ErrReason   string          `gorm:"<-:create;column:err_reason;type:varchar(128)"`
	EventAt     time.Time       `gorm:"<-:create;column:event_at;autoCreateTime:milli"`
	CreatedAt   time.Time       `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt   time.Time       `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (i *Event) TableName() string {
	return "ierc_events"
}
