package staking

import (
	"sort"

	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
	"github.com/shopspring/decimal"
)

type PoolAggregate struct {
	PoolAddress string
	Owner       string
	pools       map[uint64]*StakingPool
	//positions   map[string]map[uint64]*StakingPosition
}

func NewPoolAggregate(pool string, owner string) *PoolAggregate {
	return &PoolAggregate{
		PoolAddress: pool,
		Owner:       owner,
		pools:       make(map[uint64]*StakingPool),
		//positions:   make(map[string]map[uint64]*StakingPosition),
	}
}

func (p *PoolAggregate) InitPool(pool *StakingPool) {
	if pool.Pool != p.PoolAddress {
		return
	}

	p.pools[pool.PoolSubID] = pool
}

func (p *PoolAggregate) InitPosition(position *StakingPosition) {
	pool, existed := p.pools[position.PoolSubID]
	if !existed {
		return
	}

	pool.setPosition(position)
}

func (p *PoolAggregate) IsAdmin(poolSubID uint64, address string) bool {
	if address == p.Owner {
		return true
	}

	pool, existed := p.pools[poolSubID]
	if !existed {
		return false
	}

	return pool.IsAmin(address)
}

func (p *PoolAggregate) GetStakingPools() []*StakingPool {

	var result = make([]*StakingPool, 0, len(p.pools))
	for _, pool := range p.pools {
		result = append(result, pool)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].PoolSubID > result[j].PoolSubID
	})

	return result
}

func (p *PoolAggregate) SubPoolIsExisted(poolSubID uint64) bool {
	_, existed := p.pools[poolSubID]
	return existed
}

func (p *PoolAggregate) GetStakingPositions() []*StakingPosition {
	var positions []*StakingPosition
	for _, pool := range p.pools {
		for _, position := range pool.positions {
			positions = append(positions, position)
		}
	}

	return positions
}

func (p *PoolAggregate) UpdatePool(command *protocol.ConfigStakeCommand) error {

	if command.Owner != p.Owner {
		return protocol.NewProtocolError(protocol.StakeConfigNoPermission, "no permission")
	}

	if command.Pool != p.PoolAddress {
		return protocol.NewProtocolError(protocol.StakeConfigPoolNotMatch, "not match")
	}

	if pool, existed := p.pools[command.PoolSubID]; existed {
		return pool.UpdatePool(command)
	} else {
		pool = NewStakingPool(command)
		p.pools[pool.PoolSubID] = pool
	}

	return nil
}

func (p *PoolAggregate) Staking(blockNumber uint64, poolID uint64, staker, tick string, amount decimal.Decimal) error {

	pool, existed := p.pools[poolID]
	if !existed {
		return protocol.NewProtocolError(protocol.StakingPoolNotFound, "pool not found")
	}

	return pool.Staking(blockNumber, staker, tick, amount)
}

func (p *PoolAggregate) UnStaking(blockNumber uint64, poolID uint64, staker string, tick string, amount decimal.Decimal) error {

	pool, existed := p.pools[poolID]
	if !existed {
		return protocol.NewProtocolError(protocol.StakingPoolNotFound, "pool not found")
	}

	return pool.UnStaking(blockNumber, staker, tick, amount)
}

func (p *PoolAggregate) CanUseRewards(blockNumber uint64, staker string, amount decimal.Decimal) bool {

	var rewards = decimal.Zero
	for _, pool := range p.pools {
		rewards = rewards.Add(pool.CalcAvailableRewards(blockNumber, staker))
	}

	return rewards.GreaterThanOrEqual(amount)
}

func (p *PoolAggregate) UseRewards(blockNumber uint64, staker string, amount decimal.Decimal) error {

	useRewards := amount
	for _, pool := range p.pools {

		realUseAmount := pool.UseRewards(blockNumber, staker, useRewards)

		useRewards = useRewards.Sub(realUseAmount)
		if useRewards.IsZero() {
			return nil
		}
	}

	panic("logic error")
}
