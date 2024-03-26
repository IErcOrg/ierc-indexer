package protocol

import (
	"strconv"

	"github.com/IErcOrg/IERC_Indexer/pkg/utils"
	"github.com/shopspring/decimal"
)

// ================ deploy =================
type TokenomicsDetail struct {
	BlockNumber uint64          `json:"block_number"`
	Amount      decimal.Decimal `json:"amount"`
}

type DistributionRule struct {
	PowRatio          decimal.Decimal `json:"pow_ratio"`
	MinWorkC          string          `json:"min_work_c,omitempty"`
	DifficultyRatio   decimal.Decimal `json:"difficulty_ratio"`
	PosRatio          decimal.Decimal `json:"pos_ratio"`
	PosPool           string          `json:"pos_pool,omitempty"`
	MaxRewardBlockNum uint64          `json:"max_reward_block,omitempty"`
}

func (d *DistributionRule) TotalRatio() decimal.Decimal { return d.PowRatio.Add(d.PosRatio) }

func (d *DistributionRule) PoWPercentage() decimal.Decimal {
	return d.PowRatio.Div(d.TotalRatio())
}

func (d *DistributionRule) PoSPercentage() decimal.Decimal {
	return d.PosRatio.Div(d.TotalRatio())
}

type DeployPoWCommand struct {
	IERCTransactionBase `json:"-"`
	Tick                string             `json:"tick,omitempty"`
	Decimals            int64              `json:"decimals,omitempty"`
	MaxSupply           decimal.Decimal    `json:"max_supply"`
	TokenomicsDetails   []TokenomicsDetail `json:"tokenomics_details,omitempty"`
	DistributionRule    DistributionRule   `json:"distribution_rule"`
}

func (p *DeployPoWCommand) Validate() error {
	if err := p.IERCTransactionBase.Validate(); err != nil {
		return err
	}

	if len(p.Tick) > TickMaxLength {
		return NewProtocolError(InvalidProtocolParams, "invalid tick. length > 64")
	}

	if p.MaxSupply.LessThanOrEqual(decimal.Zero) {
		return NewProtocolError(InvalidProtocolParams, "invalid max supply")
	}

	for _, detail := range p.TokenomicsDetails {
		if detail.BlockNumber <= 0 {
			return NewProtocolError(InvalidProtocolParams, "invalid tokenomics")
		}

		if detail.Amount.LessThanOrEqual(decimal.Zero) {
			return NewProtocolError(InvalidProtocolParams, "invalid tokenomics")
		}
	}

	if !utils.IsHexAddressWith0xPrefix(p.DistributionRule.PosPool) {
		return NewProtocolError(InvalidProtocolParams, "must be hex address")
	}

	return nil
}

// ================ mint =================

type MintPoWCommand struct {
	IERCTransactionBase
	tick   string
	points decimal.Decimal
	block  uint64
	nonce  uint64
}

func NewMintPoWCommand(base IERCTransactionBase, tick string, points decimal.Decimal, block, nonce uint64) *MintPoWCommand {
	return &MintPoWCommand{
		IERCTransactionBase: base,
		tick:                tick,
		points:              points,
		block:               block,
		nonce:               nonce,
	}
}

func (m *MintPoWCommand) Tick() string            { return m.tick }
func (m *MintPoWCommand) Points() decimal.Decimal { return m.points }
func (m *MintPoWCommand) Block() uint64           { return m.block }
func (m *MintPoWCommand) Nonce() string           { return strconv.FormatUint(m.nonce, 10) }
func (m *MintPoWCommand) IsPoW() bool             { return m.block != 0 }
func (m *MintPoWCommand) IsDPoS() bool            { return m.points.GreaterThan(decimal.Zero) }

func (m *MintPoWCommand) Validate() error {
	if err := m.IERCTransactionBase.Validate(); err != nil {
		return err
	}

	if m.points.LessThan(decimal.Zero) {
		return NewProtocolError(InvalidProtocolParams, "invalid point")
	}

	if m.nonce < 0 {
		return NewProtocolError(InvalidProtocolParams, "invalid nonce")
	}

	if !m.IsPoW() && !m.IsDPoS() {
		return NewProtocolError(InvalidProtocolParams, "invalid mint")
	}

	return nil
}

type ModifyCommand struct {
	IERCTransactionBase
	Tick      string
	MaxSupply decimal.Decimal
}

type ClaimAirdropCommand struct {
	IERCTransactionBase
	Tick        string
	ClaimAmount decimal.Decimal
}
