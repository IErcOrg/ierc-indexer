package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
	jsoniter "github.com/json-iterator/go"
	"github.com/shopspring/decimal"
)

type EventKind uint8

const (
	_EventKindIERC20TickCreated EventKind = iota
	_EventKindIERC20Minted
	_EventKindIERCPoWTickCreated
	_EventKindIERCPoWMinted
	_EventKindIERC20Transferred
	_EventKindStakingPoolUpdated
)

type EventDetail interface {
	GetProtocol() protocol.Protocol
	GetOperate() protocol.Operate
	Kind() EventKind
}

type (
	IERC20TickCreated struct {
		Protocol    protocol.Protocol `json:"protocol"`
		Operate     protocol.Operate  `json:"operate"`
		Tick        string            `json:"tick"`
		Decimals    int64             `json:"decimals"`
		MaxSupply   decimal.Decimal   `json:"max_supply"`
		Limit       decimal.Decimal   `json:"limit"`
		WalletLimit decimal.Decimal   `json:"wallet_limit"`
		WorkC       string            `json:"work_c,omitempty"`
		Nonce       string            `json:"nonce"`
	}

	IERC20Minted struct {
		Protocol     protocol.Protocol `json:"protocol"`
		Operate      protocol.Operate  `json:"operate"`
		Tick         string            `json:"tick"`
		From         string            `json:"from"`
		To           string            `json:"to"`
		MintedAmount decimal.Decimal   `json:"minted_amount"`
		Gas          decimal.Decimal   `json:"gas"`
		GasPrice     decimal.Decimal   `json:"gas_price"`
		Nonce        string            `json:"nonce"`
	}

	IERCPoWTickCreated struct {
		Protocol   protocol.Protocol           `json:"protocol"`
		Operate    protocol.Operate            `json:"operate"`
		Tick       string                      `json:"tick"`
		Decimals   int64                       `json:"decimals"`
		MaxSupply  decimal.Decimal             `json:"max_supply"`
		Tokenomics []protocol.TokenomicsDetail `json:"tokenomics"`
		Rule       protocol.DistributionRule   `json:"rule"`
		Creator    string                      `json:"creator"`
	}

	IERCPoWMinted struct {
		Protocol        protocol.Protocol `json:"protocol"`
		Operate         protocol.Operate  `json:"operate"`
		Tick            string            `json:"tick"`
		From            string            `json:"from"`
		To              string            `json:"to"`
		IsPoW           bool              `json:"is_pow"`
		PoWMintedAmount decimal.Decimal   `json:"pow_minted_amount"`
		PoWTotalShare   decimal.Decimal   `json:"pow_total_share"`
		PoWMinerShare   decimal.Decimal   `json:"pow_miner_share"`
		IsPoS           bool              `json:"is_dpos"`
		PoSMintedAmount decimal.Decimal   `json:"pos_minted_amount"`
		PoSTotalShare   decimal.Decimal   `json:"pos_total_share"`
		PoSMinerShare   decimal.Decimal   `json:"pos_miner_share"`
		PoSPointsSource string            `json:"pos_points_source"`
		Gas             decimal.Decimal   `json:"gas"`
		GasPrice        decimal.Decimal   `json:"gas_price"`
		IsAirdrop       bool              `json:"is_airdrop"`
		AirdropAmount   decimal.Decimal   `json:"airdrop"`
		BurnAmount      decimal.Decimal   `json:"burn"`

		Nonce string `json:"nonce"`
	}

	IERC20Transferred struct {
		Protocol    protocol.Protocol `json:"protocol"`
		Operate     protocol.Operate  `json:"operate"`
		Tick        string            `json:"tick"`
		From        string            `json:"from"`
		To          string            `json:"to"`
		Amount      decimal.Decimal   `json:"amount"`
		EthValue    decimal.Decimal   `json:"eth_value"`
		GasPrice    decimal.Decimal   `json:"gas_price"`
		Nonce       string            `json:"nonce,omitempty"`
		SignerNonce string            `json:"signer_nonce,omitempty"`
		Sign        string            `json:"sign,omitempty"`
	}

	StakingPoolUpdated struct {
		Protocol  protocol.Protocol            `json:"protocol"`
		Operate   protocol.Operate             `json:"operate"`
		From      string                       `json:"from"`
		To        string                       `json:"to"`
		Pool      string                       `json:"pool"`
		PoolID    uint64                       `json:"pool_id"`
		Name      string                       `json:"name"`
		Owner     string                       `json:"owner"`
		Admins    []string                     `json:"admins"`
		Details   []*protocol.TickConfigDetail `json:"details"`
		StopBlock uint64                       `json:"stop_block"`
	}
)

func (i *IERC20TickCreated) GetProtocol() protocol.Protocol { return i.Protocol }
func (i *IERC20TickCreated) GetOperate() protocol.Operate   { return i.Operate }
func (i *IERC20TickCreated) Kind() EventKind                { return _EventKindIERC20TickCreated }

func (i *IERC20Minted) GetProtocol() protocol.Protocol { return i.Protocol }
func (i *IERC20Minted) GetOperate() protocol.Operate   { return i.Operate }
func (i *IERC20Minted) Kind() EventKind                { return _EventKindIERC20Minted }

func (i *IERCPoWTickCreated) GetProtocol() protocol.Protocol { return i.Protocol }
func (i *IERCPoWTickCreated) GetOperate() protocol.Operate   { return i.Operate }
func (i *IERCPoWTickCreated) Kind() EventKind                { return _EventKindIERCPoWTickCreated }

func (i *IERCPoWMinted) GetProtocol() protocol.Protocol { return i.Protocol }
func (i *IERCPoWMinted) GetOperate() protocol.Operate   { return i.Operate }
func (i *IERCPoWMinted) Kind() EventKind                { return _EventKindIERCPoWMinted }

func (i *IERC20Transferred) GetProtocol() protocol.Protocol { return i.Protocol }
func (i *IERC20Transferred) GetOperate() protocol.Operate   { return i.Operate }
func (i *IERC20Transferred) Kind() EventKind                { return _EventKindIERC20Transferred }

func (i *StakingPoolUpdated) GetProtocol() protocol.Protocol { return i.Protocol }
func (i *StakingPoolUpdated) GetOperate() protocol.Operate   { return i.Operate }
func (i *StakingPoolUpdated) Kind() EventKind                { return _EventKindStakingPoolUpdated }

var (
	_ EventDetail = (*IERC20TickCreated)(nil)
	_ EventDetail = (*IERC20Minted)(nil)
	_ EventDetail = (*IERCPoWTickCreated)(nil)
	_ EventDetail = (*IERCPoWMinted)(nil)
	_ EventDetail = (*IERC20Transferred)(nil)
	_ EventDetail = (*StakingPoolUpdated)(nil)
)

type Event interface {
	EventName() string
	GetEventKind() EventKind
	GetCurrentBlock() uint64
	GetPreviousBlock() uint64
	GetTxHash() string
	PosInIERCTxs() int
	GetProtocol() protocol.Protocol
	GetOperate() protocol.Operate
	GetErrCode() int32
	GetErrReason() string
	GetEventAt() time.Time
}

type event[T EventDetail] struct {
	BlockNumber       uint64    `json:"block_number,omitempty"`
	PrevBlockNumber   uint64    `json:"prev_block_number,omitempty"`
	TxHash            string    `json:"tx_hash,omitempty"`
	PositionInIERCTxs int       `json:"position_in_ierc_txs,omitempty"`
	From              string    `json:"from"`
	To                string    `json:"to"`
	Value             string    `json:"value"`
	Data              T         `json:"event_data,omitempty"`
	ErrCode           int32     `json:"err_code,omitempty"`
	ErrReason         string    `json:"err_reason,omitempty"`
	EventAt           time.Time `json:"event_at"`
}

func (e *event[T]) String() string {
	return fmt.Sprintf("%T(block: %d, hash: %s, data: %v)", e.Data, e.BlockNumber, e.TxHash, e.Data)
}
func (e *event[T]) EventName() string {
	return fmt.Sprintf("%T", e.Data)
}
func (e *event[T]) GetCurrentBlock() uint64        { return e.BlockNumber }
func (e *event[T]) GetPreviousBlock() uint64       { return e.PrevBlockNumber }
func (e *event[T]) GetTxHash() string              { return e.TxHash }
func (e *event[T]) PosInIERCTxs() int              { return e.PositionInIERCTxs }
func (e *event[T]) GetProtocol() protocol.Protocol { return e.Data.GetProtocol() }
func (e *event[T]) GetOperate() protocol.Operate   { return e.Data.GetOperate() }
func (e *event[T]) GetEventKind() EventKind        { return e.Data.Kind() }
func (e *event[T]) GetErrCode() int32              { return e.ErrCode }
func (e *event[T]) GetErrReason() string           { return e.ErrReason }
func (e *event[T]) GetEventAt() time.Time          { return e.EventAt }

func (e *event[T]) SetError(err error) {
	if err == nil {
		return
	}

	var pErr *protocol.ProtocolError
	if errors.As(err, &pErr) {
		e.ErrCode = pErr.Code()
		e.ErrReason = pErr.Message()
	} else {
		e.ErrCode = int32(protocol.UnknownError)
		e.ErrReason = err.Error()
	}
}

type (
	IERC20TickCreatedEvent = event[*IERC20TickCreated]
	IERC20MintedEvent      = event[*IERC20Minted]

	IERCPoWTickCreatedEvent = event[*IERCPoWTickCreated]
	IERCPoWMintedEvent      = event[*IERCPoWMinted]

	IERC20TransferredEvent = event[*IERC20Transferred]

	StakingPoolUpdatedEvent = event[*StakingPoolUpdated]
)

var (
	_ Event = (*IERC20TickCreatedEvent)(nil)
	_ Event = (*IERC20MintedEvent)(nil)
	_ Event = (*IERCPoWTickCreatedEvent)(nil)
	_ Event = (*IERCPoWMintedEvent)(nil)
	_ Event = (*IERC20TransferredEvent)(nil)
	_ Event = (*StakingPoolUpdatedEvent)(nil)
)

func NewEventFromData(kind uint8, data []byte) Event {
	switch EventKind(kind) {
	case _EventKindIERC20TickCreated:
		var event IERC20TickCreatedEvent
		_ = jsoniter.Unmarshal(data, &event)
		return &event

	case _EventKindIERC20Minted:
		var event IERC20MintedEvent
		_ = jsoniter.Unmarshal(data, &event)
		return &event

	case _EventKindIERCPoWTickCreated:
		var event IERCPoWTickCreatedEvent
		_ = jsoniter.Unmarshal(data, &event)
		return &event

	case _EventKindIERCPoWMinted:
		var event IERCPoWMintedEvent
		_ = jsoniter.Unmarshal(data, &event)
		return &event

	case _EventKindIERC20Transferred:
		var event IERC20TransferredEvent
		_ = jsoniter.Unmarshal(data, &event)
		return &event

	case _EventKindStakingPoolUpdated:
		var event StakingPoolUpdatedEvent
		_ = jsoniter.Unmarshal(data, &event)
		return &event

	default:
		panic("invalid event model")
	}
}

type EventsByBlock struct {
	BlockNumber uint64
	Events      []Event
}

func (e *EventsByBlock) CurrentBlock() uint64 {
	return e.BlockNumber
}

func (e *EventsByBlock) PreviousBlock() uint64 {
	if len(e.Events) == 0 {
		return 0
	}

	return e.Events[0].GetPreviousBlock()
}
