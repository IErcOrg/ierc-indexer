package staking

import (
	"time"

	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
	"github.com/shopspring/decimal"
)

type PositionTickDetail struct {
	Tick   string          `json:"tick"`
	Ratio  decimal.Decimal `json:"ratio"`
	Amount decimal.Decimal `json:"amount"`
}

type StakingPosition struct {
	PoolAddress      string
	PoolSubID        uint64
	Staker           string
	TickDetails      map[string]*PositionTickDetail
	RewardsPerBlock  decimal.Decimal
	Debt             decimal.Decimal
	AccReward        decimal.Decimal
	LastRewardBlock  uint64
	LastUpdatedBlock uint64
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewStakingPosition(blockNumber uint64, pool string, poolID uint64, staker string) *StakingPosition {
	return &StakingPosition{
		PoolAddress:      pool,
		PoolSubID:        poolID,
		Staker:           staker,
		TickDetails:      make(map[string]*PositionTickDetail),
		RewardsPerBlock:  decimal.Zero,
		Debt:             decimal.Zero,
		AccReward:        decimal.Zero,
		LastRewardBlock:  blockNumber,
		LastUpdatedBlock: blockNumber,
		CreatedAt:        time.Time{},
		UpdatedAt:        time.Time{},
	}
}

func (s *StakingPosition) calculateRemainingAvailableRewards() decimal.Decimal {
	return s.AccReward.Sub(s.Debt)
}

func (s *StakingPosition) calculateUnclaimedRewards(blockNumber uint64) decimal.Decimal {
	if blockNumber <= s.LastRewardBlock {
		return decimal.Zero
	}

	return s.RewardsPerBlock.Mul(decimal.NewFromInt(int64(blockNumber - s.LastRewardBlock)))
}

func (s *StakingPosition) CalcAvailableRewards(blockNumber uint64) decimal.Decimal {
	remainingRewards := s.calculateRemainingAvailableRewards()
	unclaimedRewards := s.calculateUnclaimedRewards(blockNumber)
	return remainingRewards.Add(unclaimedRewards)
}

func (s *StakingPosition) SettleRewards(blockNumber uint64) decimal.Decimal {
	if blockNumber <= s.LastRewardBlock {
		return decimal.Zero
	}

	unclaimedRewards := s.calculateUnclaimedRewards(blockNumber)
	s.AccReward = s.AccReward.Add(unclaimedRewards)
	s.LastRewardBlock = blockNumber
	return unclaimedRewards
}

func (s *StakingPosition) ResetRewardsPerBlock(blockNumber uint64, ticks map[string]*PoolTickDetail) {

	var (
		rewardsPerBlock = decimal.Zero
		details         = make(map[string]*PositionTickDetail)
	)

	for _, detail := range s.TickDetails {

		if detail.Amount.LessThanOrEqual(decimal.Zero) {
			continue
		}

		detail.Ratio = decimal.Zero
		details[detail.Tick] = detail
	}

	for _, item := range ticks {

		detail, existed := details[item.Tick]
		if !existed {

			detail = &PositionTickDetail{
				Tick:   item.Tick,
				Ratio:  item.Ratio,
				Amount: decimal.Zero,
			}
			details[item.Tick] = detail
		} else {

			detail.Ratio = item.Ratio
		}

		if detail.Ratio.LessThanOrEqual(decimal.Zero) || detail.Amount.LessThanOrEqual(decimal.Zero) {
			continue
		}

		rewardsPerBlock = rewardsPerBlock.Add(detail.Amount.Mul(detail.Ratio))
	}

	s.TickDetails = details
	s.RewardsPerBlock = rewardsPerBlock
	s.LastUpdatedBlock = blockNumber
	return
}

func (s *StakingPosition) UseRewards(blockNumber uint64, useAmount decimal.Decimal) decimal.Decimal {

	if useAmount.IsZero() {
		return useAmount
	}

	if useAmount.LessThanOrEqual(decimal.Zero) {
		panic("logic error")
	}

	available := s.calculateRemainingAvailableRewards()

	realUseAmount := decimal.Min(useAmount, available)

	s.Debt = s.Debt.Add(realUseAmount)
	s.LastUpdatedBlock = blockNumber
	return realUseAmount
}

func (s *StakingPosition) Staking(blockNumber uint64, tick string, ratio, amount decimal.Decimal) error {

	detail, existed := s.TickDetails[tick]
	if !existed {
		detail = &PositionTickDetail{
			Tick:   tick,
			Ratio:  ratio,
			Amount: decimal.Zero,
		}
		s.TickDetails[detail.Tick] = detail
	}

	detail.Amount = detail.Amount.Add(amount)
	s.RewardsPerBlock = s.RewardsPerBlock.Add(amount.Mul(ratio))
	s.LastUpdatedBlock = blockNumber
	return nil
}

func (s *StakingPosition) UnStaking(blockNumber uint64, tick string, ratio, amount decimal.Decimal) error {

	detail, existed := s.TickDetails[tick]
	if !existed {
		return protocol.NewProtocolError(protocol.UnStakingErrNoStake, "no stake")
	}

	if amount.GreaterThan(detail.Amount) {
		return protocol.NewProtocolError(protocol.UnStakingErrStakeAmountInsufficient, "insufficient stake amount")
	}

	detail.Amount = detail.Amount.Sub(amount)
	s.RewardsPerBlock = s.RewardsPerBlock.Sub(amount.Mul(ratio))
	s.LastUpdatedBlock = blockNumber
	return nil
}
