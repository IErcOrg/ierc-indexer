package balance

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type BalanceKey struct {
	Address string
	Tick    string
}

func NewBalanceKey(address, tick string) BalanceKey {
	return BalanceKey{
		Address: address,
		Tick:    tick,
	}
}

func (key *BalanceKey) String() string {
	return fmt.Sprintf("%s-%s", key.Address, key.Tick)
}

type Balance struct {
	ID               int64
	Address          string
	Tick             string
	Available        decimal.Decimal
	Freeze           decimal.Decimal
	MintedAmount     decimal.Decimal
	LastUpdatedBlock uint64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewBalance(address, tick string) *Balance {
	return &Balance{
		ID:               0,
		Address:          address,
		Tick:             tick,
		Available:        decimal.Zero,
		Freeze:           decimal.Zero,
		MintedAmount:     decimal.Zero,
		LastUpdatedBlock: 0,
		CreatedAt:        time.Time{},
		UpdatedAt:        time.Time{},
	}
}

func (entity *Balance) Key() BalanceKey {
	return NewBalanceKey(entity.Address, entity.Tick)
}

func (entity *Balance) Total() decimal.Decimal {
	return entity.Available.Add(entity.Freeze)
}

func (entity *Balance) AddAvailable(blockNumber uint64, amount decimal.Decimal) {
	entity.Available = entity.Available.Add(amount)
	entity.LastUpdatedBlock = blockNumber
}

func (entity *Balance) SubAvailable(blockNumber uint64, amount decimal.Decimal) {
	entity.Available = entity.Available.Sub(amount)
	entity.LastUpdatedBlock = blockNumber
}

func (entity *Balance) AddFreeze(blockNumber uint64, amount decimal.Decimal) {
	entity.Freeze = entity.Freeze.Add(amount)
	entity.LastUpdatedBlock = blockNumber
}

func (entity *Balance) SubFreeze(blockNumber uint64, amount decimal.Decimal) {
	entity.Freeze = entity.Freeze.Sub(amount)
	entity.LastUpdatedBlock = blockNumber
}

func (entity *Balance) AddMint(blockNumber uint64, amount decimal.Decimal) {
	entity.Available = entity.Available.Add(amount)
	entity.MintedAmount = entity.MintedAmount.Add(amount)
	entity.LastUpdatedBlock = blockNumber
}

func (entity *Balance) FreezeBalance(blockNumber uint64, amount decimal.Decimal) {
	entity.Available = entity.Available.Sub(amount)
	entity.Freeze = entity.Freeze.Add(amount)
	entity.LastUpdatedBlock = blockNumber
}

func (entity *Balance) UnfreezeBalance(blockNumber uint64, amount decimal.Decimal) {
	entity.Available = entity.Available.Add(amount)
	entity.Freeze = entity.Freeze.Sub(amount)
	entity.LastUpdatedBlock = blockNumber
}

func (entity *Balance) Marshal() ([]byte, error) {
	return json.Marshal(entity)
}

func (entity *Balance) Unmarshal(bytes []byte) error {
	return json.Unmarshal(bytes, entity)
}
