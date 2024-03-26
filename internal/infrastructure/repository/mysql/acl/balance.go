package acl

import (
	"github.com/IErcOrg/IERC_Indexer/internal/domain/balance"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/mysql/models"
)

func ConvertBalanceEntityToModel(balance *balance.Balance) *models.IERC20Balance {
	return &models.IERC20Balance{
		ID:               balance.ID,
		Address:          balance.Address,
		Tick:             balance.Tick,
		Available:        balance.Available,
		Freeze:           balance.Freeze,
		Minted:           balance.MintedAmount,
		LastUpdatedBlock: balance.LastUpdatedBlock,
		CreatedAt:        balance.CreatedAt,
		UpdatedAt:        balance.UpdatedAt,
	}
}

func ConvertBalanceModelToEntity(b *models.IERC20Balance) *balance.Balance {
	return &balance.Balance{
		ID:               b.ID,
		Address:          b.Address,
		Tick:             b.Tick,
		Available:        b.Available,
		Freeze:           b.Freeze,
		MintedAmount:     b.Minted,
		LastUpdatedBlock: b.LastUpdatedBlock,
		CreatedAt:        b.CreatedAt,
		UpdatedAt:        b.UpdatedAt,
	}
}
