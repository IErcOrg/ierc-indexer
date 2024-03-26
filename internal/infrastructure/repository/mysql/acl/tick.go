package acl

import (
	"errors"

	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
	domain "github.com/IErcOrg/IERC_Indexer/internal/domain/tick"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/mysql/models"
)

func ConvertTickEntityToModel(entity domain.Tick) *models.IERCTick {

	switch ee := entity.(type) {
	case *domain.IERC20Tick:

		data, _ := ee.Marshal()

		return &models.IERCTick{
			ID:               ee.ID,
			Protocol:         string(ee.Protocol),
			Tick:             ee.Tick,
			Decimals:         ee.Decimals,
			Creator:          ee.Creator,
			MaxSupply:        ee.MaxSupply,
			Supply:           ee.Supply,
			LastUpdatedBlock: ee.LastUpdatedAtBlock,
			Detail:           data,
			CreatedAt:        ee.CreatedAt,
			UpdatedAt:        ee.UpdatedAt,
		}

	case *domain.IERCPoWTick:
		data, _ := ee.Marshal()

		return &models.IERCTick{
			ID:               ee.ID,
			Protocol:         string(ee.Protocol),
			Tick:             ee.Tick,
			Decimals:         ee.Decimals,
			Creator:          ee.Creator,
			MaxSupply:        ee.MaxSupply,
			Supply:           ee.Supply(),
			LastUpdatedBlock: ee.LastUpdatedBlock(),
			Detail:           data,
			CreatedAt:        ee.CreatedAt,
			UpdatedAt:        ee.UpdatedAt,
		}

	default:
		panic("invalid tick")
	}
}

func ConvertTickModelToEntity(m *models.IERCTick) (domain.Tick, error) {

	switch protocol.Protocol(m.Protocol) {
	case protocol.ProtocolIERCPoW:
		entity := new(domain.IERCPoWTick)
		if err := entity.Unmarshal(m.Detail); err != nil {
			return nil, err
		}

		entity.ID = m.ID
		return entity, nil

	case protocol.ProtocolIERC20, protocol.ProtocolTERC20:
		entity := new(domain.IERC20Tick)
		if err := entity.Unmarshal(m.Detail); err != nil {
			return nil, err
		}

		entity.ID = m.ID
		return entity, nil

	default:
		return nil, errors.New("invalid protocol")
	}
}
