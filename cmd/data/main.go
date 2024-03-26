package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type BalanceItem struct {
	Address       string          `json:"address"`
	Tick          string          `json:"tick"`
	Available     decimal.Decimal `json:"available"`
	Freeze        decimal.Decimal `json:"freeze"`
	ProdAvailable decimal.Decimal `json:"prod_available"`
	ProdFreeze    decimal.Decimal `json:"prod_freeze"`
	DiffAvailable decimal.Decimal `json:"diff_available"`
}

func loadFixDataFromJsonFile(path string) []*BalanceItem {

	bytes, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	var items []*BalanceItem
	err = json.Unmarshal(bytes, &items)
	if err != nil {
		panic(err)
	}

	if len(items) != 1460 {
		panic("data missing")
	}

	return items
}

func NewDB(dsn string) *gorm.DB {

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	db.Logger = db.Logger.LogMode(logger.Silent)

	return db
}

// Balance
type Balance struct {
	Id        uint      `gorm:"column:id" form:"id" json:"id" comment:"" sql:"bigint(20),PRI"`
	Address   string    `gorm:"column:address" form:"address" json:"address" comment:"" sql:"varchar(64),MUL"`
	Tick      string    `gorm:"column:tick" form:"tick" json:"tick" comment:"" sql:"varchar(64)"`
	Balance   string    `gorm:"column:balance" form:"balance" json:"balance" comment:"" sql:"varchar(78)"`
	Hold      string    `gorm:"column:hold" form:"hold" json:"hold" comment:"" sql:"varchar(78)"`
	Decimals  int       `gorm:"column:decimals" form:"decimals" json:"decimals" comment:"" sql:"int(11)"`
	Workc     string    `gorm:"column:workc" form:"workc" json:"workc" comment:"" sql:"varchar(78)"`
	Protocol  string    `gorm:"column:protocol" form:"protocol" json:"protocol" comment:"" sql:"varchar(78)"`
	CreatedAt time.Time `gorm:"column:created_at" form:"created_at" json:"created_at,omitempty" comment:"" sql:"timestamp"`
	UpdatedAt time.Time `gorm:"column:updated_at" form:"updated_at" json:"updated_at,omitempty" comment:"" sql:"timestamp"`
}

func (m *Balance) TableName() string {
	return "balances"
}

func (m *Balance) String() string {
	data, _ := json.Marshal(m)
	return string(data)
}

func fixData(data []*BalanceItem) func(tx *gorm.DB) error {

	return func(tx *gorm.DB) error {

		for _, item := range data {
			var balance Balance
			err := tx.Table(balance.TableName()).Where("address=? and tick=?", item.Address, item.Tick).First(&balance).Error
			if err != nil {
				return err
			}

			balance.Balance = decimal.RequireFromString(balance.Balance).Add(item.DiffAvailable).String()

			err = tx.Table(balance.TableName()).Where("address=? and tick=?", item.Address, item.Tick).Updates(&balance).Error
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func main() {
	fmt.Println("fix Indexer data")
	data := loadFixDataFromJsonFile("./cmd/data/fix_data_0131.json")
	fmt.Println("loaded data number", len(data))

	db := NewDB("root:123456@(127.0.0.1:3306)/ierc_server_0131_prod?charset=utf8mb4&parseTime=True&loc=Local")

	err := db.Transaction(fixData(data))

	if err != nil {
		panic(err)
	}
}
