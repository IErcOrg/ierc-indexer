package staking

import (
	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
	"github.com/shopspring/decimal"
)

type PoolTickDetail struct {
	Index         int             `json:"idx"`
	Tick          string          `json:"tick"`
	Ratio         decimal.Decimal `json:"ratio"`
	Amount        decimal.Decimal `json:"amount"`
	MaxAmount     decimal.Decimal `json:"max_amount"`
	HistoryAmount decimal.Decimal `json:"history_amount"`
}

func (s *PoolTickDetail) Copy() *PoolTickDetail {
	return &PoolTickDetail{
		Index:         s.Index,
		Tick:          s.Tick,
		Ratio:         s.Ratio.Copy(),
		Amount:        s.Amount.Copy(),
		MaxAmount:     s.MaxAmount.Copy(),
		HistoryAmount: s.HistoryAmount.Copy(),
	}
}

type StakingPoolDetail struct {
	Name        string                     `json:"name"`
	Owner       string                     `json:"owner"`
	Admins      []string                   `json:"admins,omitempty"`
	StartBlock  uint64                     `json:"start_block"`
	StopBlock   uint64                     `json:"stop_block,omitempty"`
	TickDetails map[string]*PoolTickDetail `json:"details,omitempty"`
}

type StakingPool struct {
	Pool             string            `json:"pool,omitempty"`
	PoolSubID        uint64            `json:"poolSubID,omitempty"`
	Detail           StakingPoolDetail `json:"detail"`
	LastUpdatedBlock uint64            `json:"lastUpdatedBlock,omitempty"`

	positions map[string]*StakingPosition
}

func NewStakingPool(command *protocol.ConfigStakeCommand) *StakingPool {

	var details = make(map[string]*PoolTickDetail)
	for idx, tickWithRatio := range command.Details {
		details[tickWithRatio.Tick] = &PoolTickDetail{
			Index:     idx,
			Tick:      tickWithRatio.Tick,
			Ratio:     tickWithRatio.RewardsRatioPerBlock,
			MaxAmount: tickWithRatio.MaxAmount,
			Amount:    decimal.Zero,
		}
	}

	return &StakingPool{
		Pool:      command.Pool,
		PoolSubID: command.PoolSubID,
		Detail: StakingPoolDetail{
			Name:        command.Name,
			Owner:       command.Owner,
			Admins:      command.Admins,
			StartBlock:  command.BlockNumber,
			StopBlock:   command.StopBlock,
			TickDetails: details,
		},
		LastUpdatedBlock: command.BlockNumber,
		positions:        make(map[string]*StakingPosition),
	}
}

func (p *StakingPool) getPosition(staker string) *StakingPosition {
	if p.positions == nil {
		return nil
	}

	position, existed := p.positions[staker]
	if !existed {
		return nil
	}

	return position
}

func (p *StakingPool) setPosition(position *StakingPosition) {
	if p.positions == nil {
		p.positions = make(map[string]*StakingPosition)
	}

	p.positions[position.Staker] = position
}

func (p *StakingPool) IsAmin(address string) bool {
	if address == p.Detail.Owner {
		return true
	}

	for _, admin := range p.Detail.Admins {
		if admin == address {
			return true
		}
	}

	return false
}

func (p *StakingPool) IsTimeLimited() bool {
	return p.Detail.StopBlock != 0
}

func (p *StakingPool) IsEnd(currBlock uint64) bool {
	return p.Detail.StopBlock != 0 && p.Detail.StopBlock < currBlock
}

func (p *StakingPool) CanStaking(blockNumber uint64, tick string, amount decimal.Decimal) error {

	detail, existed := p.Detail.TickDetails[tick]
	if !existed {
		return protocol.NewProtocolError(protocol.StakingTickUnsupported, "tick unsupported")
	}

	if detail.Ratio.IsZero() {
		return protocol.NewProtocolError(protocol.StakingTickUnsupported, "tick unsupported")
	}

	if p.IsTimeLimited() {

		if blockNumber >= p.Detail.StopBlock {
			return protocol.NewProtocolError(protocol.StakingPoolAlreadyStopped, "pool already stopped")
		}

		if detail.MaxAmount.Sub(detail.Amount).LessThan(amount) {
			return protocol.NewProtocolError(protocol.StakingPoolIsFulled, "pool is fulled")
		}
	}

	return nil
}

func (p *StakingPool) CanUnStaking(blockNumber uint64, tick string, amount decimal.Decimal) error {

	detail, existed := p.Detail.TickDetails[tick]
	if !existed {
		return protocol.NewProtocolError(protocol.StakingTickUnsupported, "tick unsupported")
	}

	if detail.Amount.LessThan(amount) {
		return protocol.NewProtocolError(protocol.UnStakingErrStakeAmountInsufficient, "invalid amount")
	}

	if p.IsTimeLimited() {
		if blockNumber <= p.Detail.StopBlock {
			return protocol.NewProtocolError(protocol.UnStakingErrNotYetUnlocked, "not yet unlocked")
		}
	}

	return nil
}

func (p *StakingPool) CalcAvailableRewards(blockNumber uint64, staker string) decimal.Decimal {
	position := p.getPosition(staker)
	if position == nil {
		return decimal.Zero
	}

	if p.IsTimeLimited() {
		blockNumber = min(blockNumber, p.Detail.StopBlock)
	}

	return position.CalcAvailableRewards(blockNumber)
}

func (p *StakingPool) UpdatePool(command *protocol.ConfigStakeCommand) error {
	if p.IsTimeLimited() {
		return p.updateTimeLimitedPool(command)
	} else {
		return p.updateUnlimitedPool(command)
	}
}

func (p *StakingPool) updateUnlimitedPool(command *protocol.ConfigStakeCommand) error {

	tickDetails := make(map[string]*PoolTickDetail)
	for _, info := range p.Detail.TickDetails {

		if info.Amount.LessThanOrEqual(decimal.Zero) {
			continue
		}

		info.Ratio = decimal.Zero
		tickDetails[info.Tick] = info
	}

	for idx, item := range command.Details {
		tick, existed := tickDetails[item.Tick]
		if !existed {
			tickDetails[item.Tick] = &PoolTickDetail{
				Index:  idx,
				Tick:   item.Tick,
				Ratio:  item.RewardsRatioPerBlock,
				Amount: decimal.Zero,
			}
		} else {
			tick.Index = idx
			tick.Ratio = item.RewardsRatioPerBlock
		}
	}

	for _, position := range p.positions {

		p.settleRewards(command.BlockNumber, position)

		position.ResetRewardsPerBlock(command.BlockNumber, tickDetails)
	}

	p.Detail.Name = command.Name
	p.Detail.TickDetails = tickDetails
	p.Detail.Admins = command.Admins
	p.LastUpdatedBlock = command.BlockNumber

	return nil
}

func (p *StakingPool) updateTimeLimitedPool(command *protocol.ConfigStakeCommand) error {

	if p.IsEnd(command.BlockNumber) {
		return protocol.NewProtocolError(protocol.StakingPoolIsEnded, "pool is ended")
	}

	tickDetails := make(map[string]*PoolTickDetail)
	for _, info := range p.Detail.TickDetails {

		if info.Amount.LessThanOrEqual(decimal.Zero) {
			continue
		}

		info.Ratio = decimal.Zero
		info.MaxAmount = decimal.Zero
		tickDetails[info.Tick] = info
	}

	for idx, item := range command.Details {
		tick, existed := tickDetails[item.Tick]
		if !existed {
			tickDetails[item.Tick] = &PoolTickDetail{
				Index:     idx,
				Tick:      item.Tick,
				Ratio:     item.RewardsRatioPerBlock,
				MaxAmount: item.MaxAmount,
				Amount:    decimal.Zero,
			}
		} else {

			if item.MaxAmount.LessThan(tick.Amount) {
				return protocol.NewProtocolError(protocol.StakingPoolMaxAmountLessThanCurrentAmount, "max amount less than current amount")
			}

			tick.Index = idx
			tick.Ratio = item.RewardsRatioPerBlock
			tick.MaxAmount = item.MaxAmount
		}
	}

	for _, position := range p.positions {

		p.settleRewards(command.BlockNumber, position)

		position.ResetRewardsPerBlock(command.BlockNumber, tickDetails)
	}

	p.Detail.Name = command.Name
	p.Detail.TickDetails = tickDetails
	p.Detail.Admins = command.Admins
	p.LastUpdatedBlock = command.BlockNumber

	return nil
}

func (p *StakingPool) Staking(blockNumber uint64, staker, tick string, amount decimal.Decimal) error {

	if err := p.CanStaking(blockNumber, tick, amount); err != nil {
		return err
	}

	detail, _ := p.Detail.TickDetails[tick]

	position := p.getPosition(staker)
	if position == nil {
		position = NewStakingPosition(blockNumber, p.Pool, p.PoolSubID, staker)
		p.setPosition(position)
	} else {
		p.settleRewards(blockNumber, position)
	}

	_ = position.Staking(blockNumber, detail.Tick, detail.Ratio, amount)

	detail.Amount = detail.Amount.Add(amount)
	if p.IsTimeLimited() {
		detail.HistoryAmount = detail.HistoryAmount.Add(amount)
	}
	p.LastUpdatedBlock = blockNumber

	return nil
}

func (p *StakingPool) UnStaking(blockNumber uint64, staker string, tick string, amount decimal.Decimal) error {

	if err := p.CanUnStaking(blockNumber, tick, amount); err != nil {
		return err
	}

	detail, _ := p.Detail.TickDetails[tick]

	position := p.getPosition(staker)
	if position == nil {
		return protocol.NewProtocolError(protocol.UnStakingErrNoStake, "no stake")
	}

	p.settleRewards(blockNumber, position)

	if err := position.UnStaking(blockNumber, detail.Tick, detail.Ratio, amount); err != nil {
		return err
	}

	detail.Amount = detail.Amount.Sub(amount)
	p.LastUpdatedBlock = blockNumber

	return nil
}

func (p *StakingPool) UseRewards(blockNumber uint64, staker string, amount decimal.Decimal) decimal.Decimal {

	position := p.getPosition(staker)
	if position == nil {
		return decimal.Zero
	}

	p.settleRewards(blockNumber, position)

	return position.UseRewards(blockNumber, amount)
}

func (p *StakingPool) settleRewards(blockNumber uint64, position *StakingPosition) {

	if p.IsTimeLimited() {
		blockNumber = min(blockNumber, p.Detail.StopBlock)
	}

	position.SettleRewards(blockNumber)
}
