package acl

import (
	"time"

	"github.com/IErcOrg/IERC_Indexer/internal/domain"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/mysql/models"
	jsoniter "github.com/json-iterator/go"
	"github.com/shopspring/decimal"
)

// ============ event

func ConvertEventToModel(entity domain.Event) *models.Event {
	var (
		ethFrom  string
		ethTo    string
		iercFrom string
		iercTo   string
		tick     string
		amount   decimal.Decimal
		sign     string
	)

	switch ee := entity.(type) {
	case *domain.IERC20TickCreatedEvent:
		ethFrom = ee.From
		ethTo = ee.To
		iercFrom = ee.From
		iercTo = ee.To
		tick = ee.Data.Tick

	case *domain.IERC20MintedEvent:
		ethFrom = ee.From
		ethTo = ee.To
		iercFrom = ee.Data.From
		iercTo = ee.Data.To
		tick = ee.Data.Tick
		amount = ee.Data.MintedAmount

	case *domain.IERCPoWTickCreatedEvent:
		ethFrom = ee.From
		ethTo = ee.To
		iercFrom = ee.From
		iercTo = ee.To
		tick = ee.Data.Tick

	case *domain.IERCPoWMintedEvent:
		ethFrom = ee.From
		ethTo = ee.To
		iercFrom = ee.Data.From
		iercTo = ee.Data.To
		tick = ee.Data.Tick
		amount = ee.Data.PoSMintedAmount.Add(ee.Data.PoWMintedAmount)

	case *domain.IERC20TransferredEvent:
		ethFrom = ee.From
		ethTo = ee.To
		iercFrom = ee.Data.From
		iercTo = ee.Data.To
		sign = ee.Data.Sign
		tick = ee.Data.Tick
		amount = ee.Data.Amount

	case *domain.StakingPoolUpdatedEvent:
		ethFrom = ee.From
		ethTo = ee.To
		iercFrom = ee.From
		iercTo = ee.To

	default:
		panic("invalid event type")
	}

	data, _ := jsoniter.Marshal(entity)
	return &models.Event{
		ID:          0,
		BlockNumber: entity.GetCurrentBlock(),
		TxHash:      entity.GetTxHash(),
		Operate:     string(entity.GetOperate()),
		Tick:        tick,
		ETHFrom:     ethFrom,
		ETHTo:       ethTo,
		IERCFrom:    iercFrom,
		IERCTo:      iercTo,
		Amount:      amount,
		Sign:        sign,
		EventKind:   uint8(entity.GetEventKind()),
		Event:       data,
		ErrCode:     entity.GetErrCode(),
		ErrReason:   entity.GetErrReason(),
		EventAt:     entity.GetEventAt(),
		CreatedAt:   time.Time{},
		UpdatedAt:   time.Time{},
	}
}

func ConvertModelToEvent(m *models.Event) domain.Event {
	return domain.NewEventFromData(m.EventKind, m.Event)
}
