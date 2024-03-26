package tick

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
	"github.com/shopspring/decimal"
)

const Ethpi = "ethpi"

var (
	ErrNoPermission                     = errors.New("no permission")
	ErrInvalidAmount                    = errors.New("invalid amount")
	ErrMaxAmountLessThanSupply          = errors.New("max amount less than supply")
	ErrAirdropAmountExceedsRemainSupply = errors.New("claim amount exceeds remain supply")
)

var MinDecimal = decimal.NewFromInt(1).Shift(-18)

type IERCPoWTick struct {
	ID            int64                       `json:"id,omitempty"`
	Tick          string                      `json:"tick,omitempty"`
	Protocol      protocol.Protocol           `json:"protocol,omitempty"`
	Decimals      int64                       `json:"decimals,omitempty"`
	Tokenomics    []protocol.TokenomicsDetail `json:"tokenomics,omitempty"`
	Rule          protocol.DistributionRule   `json:"rule"`
	MaxSupply     decimal.Decimal             `json:"max_supply"`
	AirdropAmount decimal.Decimal             `json:"airdrop_amount"`

	PoWSupply     decimal.Decimal `json:"pow_supply"`
	PoWLastBlock  uint64          `json:"pow_last_block"`
	PoWBurnAmount decimal.Decimal `json:"pow_burn_amount"`

	PoSSupply     decimal.Decimal `json:"pos_supply"`
	PoSLastBlock  uint64          `json:"pos_last_block"`
	PoSBurnAmount decimal.Decimal `json:"pos_burn_amount"`

	LastUpdateBlock uint64    `json:"last_update_block"`
	Creator         string    `json:"creator"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	powCanMint decimal.Decimal
	powBurn    decimal.Decimal
	posCanMint decimal.Decimal
	posBurn    decimal.Decimal

	powRemainCanMint decimal.Decimal
	posRemainCanMint decimal.Decimal
}

func NewIERCPoWTickFromDeployCommand(command *protocol.DeployPoWCommand) *IERCPoWTick {

	return &IERCPoWTick{
		ID:            0,
		Tick:          command.Tick,
		Protocol:      command.Protocol,
		Decimals:      command.Decimals,
		Tokenomics:    command.TokenomicsDetails,
		Rule:          command.DistributionRule,
		MaxSupply:     command.MaxSupply,
		PoWSupply:     decimal.Zero,
		PoWLastBlock:  command.BlockNumber,
		PoWBurnAmount: decimal.Zero,
		PoSSupply:     decimal.Zero,
		PoSLastBlock:  command.BlockNumber,
		PoSBurnAmount: decimal.Zero,
		Creator:       command.From,
		CreatedAt:     command.EventAt,
		UpdatedAt:     command.EventAt,

		powCanMint: decimal.Zero,
		powBurn:    decimal.Zero,
		posCanMint: decimal.Zero,
		posBurn:    decimal.Zero,
	}
}

func (entity *IERCPoWTick) GetID() int64                   { return entity.ID }
func (entity *IERCPoWTick) GetName() string                { return entity.Tick }
func (entity *IERCPoWTick) GetProtocol() protocol.Protocol { return entity.Protocol }
func (entity *IERCPoWTick) LastUpdatedBlock() uint64 {
	return max(entity.PoWLastBlock, entity.PoSLastBlock, entity.LastUpdateBlock)
}

func (entity *IERCPoWTick) Supply() decimal.Decimal {
	return entity.PoWSupply.Add(entity.PoSSupply).Add(entity.AirdropAmount)
}

func (entity *IERCPoWTick) remainSupply() decimal.Decimal {
	return entity.MaxSupply.Sub(entity.Supply())
}

func (entity *IERCPoWTick) PoWRemainSupply() decimal.Decimal {
	powMaxSupply := entity.MaxSupply.Sub(entity.AirdropAmount).Mul(entity.Rule.PoWPercentage())
	return decimal.Max(powMaxSupply.Sub(entity.PoWSupply), decimal.Zero)
}

func (entity *IERCPoWTick) PoSRemainSupply() decimal.Decimal {
	posMaxSupply := entity.MaxSupply.Sub(entity.AirdropAmount).Mul(entity.Rule.PoSPercentage())
	return decimal.Max(posMaxSupply.Sub(entity.PoSSupply), decimal.Zero)
}

func (entity *IERCPoWTick) CalcSupplyByBlockNumber(blockNumber uint64) decimal.Decimal {
	var supply = entity.Supply()

	if blockNumber == entity.PoWLastBlock {
		supply = supply.Add(entity.powRemainCanMint).Add(entity.powBurn)
	} else {
		powCanMint, powBurnAmount := entity.calcCanMintAndBurnAmount(
			entity.PoWLastBlock,
			blockNumber,
			entity.Rule.PoWPercentage(),
			entity.PoWRemainSupply(),
			entity.getRewardBlockNum(blockNumber, true),
		)
		supply = supply.Add(powCanMint).Add(powBurnAmount)
	}

	if blockNumber == entity.PoSLastBlock {
		supply = supply.Add(entity.posRemainCanMint).Add(entity.posBurn)
	} else {
		posCanMint, posBurnAmount := entity.calcCanMintAndBurnAmount(
			entity.PoSLastBlock,
			blockNumber,
			entity.Rule.PoSPercentage(),
			entity.PoSRemainSupply(),
			entity.getRewardBlockNum(blockNumber, false),
		)
		supply = supply.Add(posCanMint).Add(posBurnAmount)
	}

	return supply
}

func (entity *IERCPoWTick) CalculateMintShareBasedOnHash(blockNumber uint64, hash string) decimal.Decimal {

	currDifficulty := countLeadingZeros(hash)
	minDifficulty := countLeadingZeros(entity.Rule.MinWorkC)

	if currDifficulty < minDifficulty {
		return decimal.Zero
	}

	if entity.Tick == Ethpi {

		switch {

		case blockNumber > protocol.PoWMintLimitBlockHeight:
			return decimal.NewFromInt(1)

		case blockNumber > protocol.DPoSMintMintPointsLimitBlockHeight:
			return decimal.NewFromInt(5).Pow(decimal.NewFromInt(int64(currDifficulty - minDifficulty)))

		default:
			return entity.Rule.DifficultyRatio.Pow(decimal.NewFromInt(int64(currDifficulty - minDifficulty)))
		}

	} else {
		return entity.Rule.DifficultyRatio.Pow(decimal.NewFromInt(int64(currDifficulty - minDifficulty)))
	}
}

func (entity *IERCPoWTick) getRewardBlockNum(blockNumber uint64, isPoW bool) uint64 {
	if entity.Tick != Ethpi {
		return entity.Rule.MaxRewardBlockNum
	}

	if !isPoW {
		return entity.Rule.MaxRewardBlockNum
	}

	switch {

	case blockNumber > protocol.PoWMintLimitBlockHeight:
		return 2

	default:
		return entity.Rule.MaxRewardBlockNum
	}
}

type PoWMintParams struct {
	CurrentBlock   uint64
	EffectiveBlock uint64
	IsPoW          bool
	IsDPoS         bool
	TotalPoWShare  decimal.Decimal
	MinerPoWShare  decimal.Decimal
	TotalPoSShare  decimal.Decimal
	MinerPoSShare  decimal.Decimal
}

func (p *PoWMintParams) String() string {
	data, _ := json.Marshal(p)
	return fmt.Sprintf("%s", data)
}

func (p *PoWMintParams) Validate() error {
	if p.MinerPoWShare.GreaterThan(p.TotalPoWShare) {
		return errors.New("miner share > total share")
	}

	if p.MinerPoSShare.GreaterThan(p.TotalPoSShare) {
		return errors.New("miner share > total share")
	}

	return nil
}

func (entity *IERCPoWTick) CanMint(params *PoWMintParams) error {

	switch {
	case params.IsPoW && params.IsDPoS:

		if params.MinerPoWShare.IsZero() && params.MinerPoSShare.IsZero() {
			return protocol.NewProtocolError(protocol.MintErr, "invalid mint")
		}

		if entity.remainSupply().LessThanOrEqual(MinDecimal) {
			return protocol.NewProtocolError(protocol.MintAmountExceedLimit, "already mint done")
		}

		var diff = max(params.EffectiveBlock, params.CurrentBlock) - min(params.EffectiveBlock, params.CurrentBlock)
		if diff > 5 {
			return protocol.NewProtocolError(protocol.MintBlockExpires, "block expires")
		}

	case params.IsPoW:

		if params.MinerPoWShare.IsZero() {
			return protocol.NewProtocolError(protocol.MintPoWInvalidHash, "invalid hash")
		}

		var diff = max(params.EffectiveBlock, params.CurrentBlock) - min(params.EffectiveBlock, params.CurrentBlock)
		if diff > 5 {
			return protocol.NewProtocolError(protocol.MintBlockExpires, "block expires")
		}

		if entity.PoWRemainSupply().LessThanOrEqual(MinDecimal) {
			return protocol.NewProtocolError(protocol.MintAmountExceedLimit, "pow already mint done")
		}

	case params.IsDPoS:

		if params.MinerPoSShare.IsZero() {
			return protocol.NewProtocolError(protocol.MintPoSInvalidShare, "invalid points")
		}

		if entity.PoSRemainSupply().LessThanOrEqual(MinDecimal) {
			return protocol.NewProtocolError(protocol.MintAmountExceedLimit, "pos already mint done")
		}

	default:
		return protocol.NewProtocolError(protocol.MintErr, "invalid mint")
	}

	return nil
}

func (entity *IERCPoWTick) calcMintStartBlock() uint64 {
	var block uint64 = math.MaxUint64
	for _, tokenomic := range entity.Tokenomics {
		block = min(block, tokenomic.BlockNumber)
	}

	return block
}

func (entity *IERCPoWTick) calcOutputAmountOfBlock(targetBlock uint64) decimal.Decimal {
	for i := len(entity.Tokenomics) - 1; i >= 0; i-- {
		tokenomics := entity.Tokenomics[i]
		if targetBlock >= tokenomics.BlockNumber {
			return tokenomics.Amount
		}
	}

	return decimal.Zero
}

// input:
// - startBlock
// - targetBlock
// - outputRatio
// - remainSupply
//
// output:
// - outputAmount
// - burnAmount
func (entity *IERCPoWTick) calcCanMintAndBurnAmount(startBlock, targetBlock uint64, outputRatio, remainSupply decimal.Decimal, maxRewardBlockNum uint64) (decimal.Decimal, decimal.Decimal) {

	var canMintAmount = decimal.Zero
	var burnAmount = decimal.Zero

	var mintStartBlock = entity.calcMintStartBlock()
	startBlock = max(startBlock, mintStartBlock)

	for i := startBlock + 1; i <= targetBlock; i++ {

		outputOfTargetBlock := entity.calcOutputAmountOfBlock(i)

		realOutputAmount := decimal.Min(outputOfTargetBlock.Mul(outputRatio), remainSupply)

		if maxRewardBlockNum > 0 {
			if i <= startBlock+maxRewardBlockNum {
				canMintAmount = canMintAmount.Add(realOutputAmount)
			} else {
				burnAmount = burnAmount.Add(realOutputAmount)
			}
		} else {
			canMintAmount = canMintAmount.Add(realOutputAmount)
		}

		remainSupply = remainSupply.Sub(realOutputAmount)
		if remainSupply.IsZero() {
			break
		}
	}

	return canMintAmount, burnAmount
}

func (entity *IERCPoWTick) updateCanMintAndBurnAmount(params *PoWMintParams) {
	if params.IsPoW && params.CurrentBlock > entity.PoWLastBlock {
		powCanMint, powBurnAmount := entity.calcCanMintAndBurnAmount(
			entity.PoWLastBlock,
			params.CurrentBlock,
			entity.Rule.PoWPercentage(),
			entity.PoWRemainSupply(),
			entity.getRewardBlockNum(params.CurrentBlock, true),
		)
		entity.powCanMint = powCanMint
		entity.powRemainCanMint = powCanMint
		entity.powBurn = powBurnAmount
		entity.PoWLastBlock = params.CurrentBlock
	}

	if params.IsDPoS && params.CurrentBlock > entity.PoSLastBlock {
		posCanMint, posBurnAmount := entity.calcCanMintAndBurnAmount(
			entity.PoSLastBlock,
			params.CurrentBlock,
			entity.Rule.PoSPercentage(),
			entity.PoSRemainSupply(),
			entity.getRewardBlockNum(params.CurrentBlock, false),
		)
		entity.posCanMint = posCanMint
		entity.posRemainCanMint = posCanMint
		entity.posBurn = posBurnAmount
		entity.PoSLastBlock = params.CurrentBlock
	}
}

func (entity *IERCPoWTick) Mint(params *PoWMintParams) (decimal.Decimal, decimal.Decimal) {

	entity.updateCanMintAndBurnAmount(params)

	var powMintAmount, posMintAmount decimal.Decimal

	if !params.TotalPoWShare.IsZero() && !params.MinerPoWShare.IsZero() {
		powMintAmount = entity.powCanMint.Mul(params.MinerPoWShare).Div(params.TotalPoWShare) // canMint * minerShare / totalShare
		powMintAmount = powMintAmount.RoundFloor(18)
		entity.powRemainCanMint = entity.powRemainCanMint.Sub(powMintAmount)
		entity.PoWSupply = entity.PoWSupply.Add(powMintAmount)
	}

	if !params.TotalPoSShare.IsZero() && !params.MinerPoSShare.IsZero() {
		posMintAmount = entity.posCanMint.Mul(params.MinerPoSShare).Div(params.TotalPoSShare) // canMint * minerShare / totalShare
		posMintAmount = posMintAmount.RoundFloor(18)
		entity.posRemainCanMint = entity.posRemainCanMint.Sub(posMintAmount)
		entity.PoSSupply = entity.PoSSupply.Add(posMintAmount)
	}

	return powMintAmount, posMintAmount
}

func (entity *IERCPoWTick) UpdateMaxSupply(blockNumber uint64, creator string, amount decimal.Decimal) error {

	if entity.Creator != creator {
		return ErrNoPermission
	}

	var supply = entity.CalcSupplyByBlockNumber(blockNumber)

	if amount.LessThan(supply) {
		return ErrMaxAmountLessThanSupply
	}

	entity.MaxSupply = amount
	entity.LastUpdateBlock = blockNumber
	return nil
}

func (entity *IERCPoWTick) ClaimAirdrop(blockNumber uint64, receiver string, amount decimal.Decimal) error {

	if entity.Creator != receiver {
		return ErrNoPermission
	}

	var supply = entity.CalcSupplyByBlockNumber(blockNumber)
	var remainSupply = entity.MaxSupply.Sub(supply)

	if amount.LessThanOrEqual(decimal.Zero) {
		return ErrInvalidAmount
	}

	if amount.GreaterThan(remainSupply) {
		return ErrAirdropAmountExceedsRemainSupply
	}

	entity.AirdropAmount = entity.AirdropAmount.Add(amount)
	entity.LastUpdateBlock = blockNumber
	return nil
}

func (entity *IERCPoWTick) Burn() (decimal.Decimal, decimal.Decimal) {

	var powBurn, posBurn decimal.Decimal

	if entity.powBurn.GreaterThan(decimal.Zero) {
		powBurn = entity.powBurn.Copy()
		entity.PoWBurnAmount = entity.PoWBurnAmount.Add(powBurn)
		entity.PoWSupply = entity.PoWSupply.Add(powBurn)
		entity.powBurn = decimal.Zero
	}

	if entity.posBurn.GreaterThan(decimal.Zero) {
		posBurn = entity.posBurn.Copy()
		entity.PoSBurnAmount = entity.PoSBurnAmount.Add(posBurn)
		entity.PoSSupply = entity.PoSSupply.Add(posBurn)
		entity.posBurn = decimal.Zero
	}

	return powBurn, posBurn
}

func countLeadingZeros(hash string) int {

	hash = strings.TrimPrefix(hash, "0x")

	count := 0
	for _, char := range hash {
		if char == '0' {
			count++
		} else {
			break
		}
	}

	return count
}

func (entity *IERCPoWTick) Marshal() ([]byte, error) {
	return json.Marshal(entity)
}

func (entity *IERCPoWTick) Unmarshal(bytes []byte) error {
	return json.Unmarshal(bytes, entity)
}
