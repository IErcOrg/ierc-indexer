package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Block struct {
	ID               int64     `gorm:"<-:create;column:id;primaryKey;autoIncrement"`
	Number           uint64    `gorm:"<-:create;column:block_number;type:bigint;uniqueIndex:uni_block_number"`
	Hash             string    `gorm:"<-:create;column:block_hash;type:varchar(66);not null"`
	ParentHash       string    `gorm:"<-:create;column:parent_hash;type:varchar(66);not null;default:''"`
	TransactionCount int       `gorm:"<-:create;column:tx_count;type:bigint;index:idx_count;not null;default:0"`
	IsProcessed      bool      `gorm:"column:is_processed;type:int;not null;default:0"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (b *Block) TableName() string {
	return "blocks"
}

type Transaction struct {
	ID            int64           `gorm:"<-:create;column:id;primaryKey;autoIncrement"`
	BlockNumber   uint64          `gorm:"<-:create;column:block_number;type:bigint;uniqueIndex:uni_num_pos,priority:1"`
	PositionInTxs int64           `gorm:"<-:create;column:position;type:bigint;uniqueIndex:uni_num_pos,priority:2"`
	Hash          string          `gorm:"<-:create;column:hash;type:varchar(66);index:idx_hash;not null"`
	From          string          `gorm:"<-:create;column:from;type:varchar(42);index:idx_from;not null"`
	To            string          `gorm:"<-:create;column:to;type:varchar(42);index:idx_to;not null"`
	Value         decimal.Decimal `gorm:"<-:create;column:value;type:decimal(65,0);not null;default:0"`
	Gas           decimal.Decimal `gorm:"<-:create;column:gas;type:decimal(65,0);not null;default:0"`
	GasPrice      decimal.Decimal `gorm:"<-:create;column:gas_price;type:decimal(65,0);not null;default:0"`
	Data          string          `gorm:"<-:create;column:data;type:MEDIUMBLOB;not null"`
	Nonce         uint64          `gorm:"<-:create;column:nonce;type:int;not null;comment:'Nonce'"`

	IsProcessed bool      `gorm:"column:is_processed;type:int;not null;default:0"`
	Code        int32     `gorm:"column:code;type:int;not null;default:0"`
	Remark      string    `gorm:"column:remark;type:varchar(128)"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime:milli"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime:milli"`
}

func (t *Transaction) TableName() string {
	return "ierc_transactions"
}
