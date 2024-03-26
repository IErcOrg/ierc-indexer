package service

import (
	"context"
	"time"

	"github.com/IErcOrg/IERC_Indexer/internal/conf"
	"github.com/IErcOrg/IERC_Indexer/internal/domain"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/balance"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/staking"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/tick"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
)

type BlockService struct {
	logger          *log.Helper
	blockRepo       domain.BlockRepository
	eventRepo       domain.EventRepository
	transactionRepo domain.TransactionRepository
	tickRepo        tick.TickRepository
	balanceRepo     balance.BalanceRepository
	stakingRepo     staking.StakingRepository

	// config
	invalidHashMap map[string]struct{}
	feeStartBlock  uint64

	// runtime
	lastHandleBlock uint64
}

func NewBlockService(
	c *conf.Config,
	logger log.Logger,
	blockRepo domain.BlockRepository,
	eventRepo domain.EventRepository,
	transactionRepo domain.TransactionRepository,
	tickRepo tick.TickRepository,
	balanceRepo balance.BalanceRepository,
	stakingRepo staking.StakingRepository,
) (*BlockService, error) {
	lastBlock, err := eventRepo.GetBlockNumberByLastEvent(context.Background())
	if err != nil {
		return nil, err
	}

	return &BlockService{
		logger:          log.NewHelper(log.With(logger, "module", "BlockService")),
		blockRepo:       blockRepo,
		eventRepo:       eventRepo,
		transactionRepo: transactionRepo,
		tickRepo:        tickRepo,
		balanceRepo:     balanceRepo,
		stakingRepo:     stakingRepo,
		invalidHashMap:  c.InvalidTxHash,
		feeStartBlock:   c.Runtime.GetFeeStartBlock(),
		lastHandleBlock: lastBlock,
	}, nil
}

func (b *BlockService) GetLastHandleBlock() uint64 {
	return b.lastHandleBlock
}

func (b *BlockService) SyncBlock(ctx context.Context, blocks []*domain.Block) error {
	return b.blockRepo.BulkSaveBlock(ctx, blocks)
}

func (b *BlockService) HandleBlock(ctx context.Context, block *domain.Block) error {
	b.logger.Infof("start handle block. block_number: %d, transaction: %d", block.Number, len(block.Transactions))
	var (
		start      = time.Now()
		eventCount int
	)
	defer func() {
		b.logger.Infof("handle block done. block_number: %d, events: %d, duration: %v", block.Number, eventCount, time.Since(start))
	}()

	aggregate, err := b.preprocessing(ctx, block)
	if err != nil {
		return err
	}

	aggregate.Handle()

	if err := b.saveToDBWithTx(ctx, aggregate); err != nil {
		return err
	}

	eventCount = len(aggregate.Events)
	if len(aggregate.Events) != 0 {
		b.lastHandleBlock = aggregate.Block.Number
	}

	return nil
}

func (b *BlockService) preprocessing(ctx context.Context, block *domain.Block) (*domain.AggregateRoot, error) {

	var (
		aggregate = domain.NewBlockAggregate(b.lastHandleBlock, block, b.invalidHashMap, b.feeStartBlock)

		tickSet         = mapset.NewSet[string]()
		balanceSet      = mapset.NewSet[balance.BalanceKey]()
		signatureSet    = mapset.NewSet[string]()
		unfreezeSignSet = mapset.NewSet[string]()
	)

	pools, err := b.stakingRepo.LoadAllPools(ctx)
	if err != nil {
		return nil, err
	}

	aggregate.StakingPools = pools

loop:
	for _, transaction := range block.Transactions {

		if transaction.IERCTransaction == nil {
			//b.logger.Debugf("ignore transactions that are not IERC20. tx: %v", transaction)
			transaction.IsProcessed = true
			continue loop
		}

		if err := transaction.IERCTransaction.Validate(); err != nil {
			transaction.Code = err.(*protocol.ProtocolError).Code()
			transaction.Remark = err.Error()
			transaction.IsProcessed = true
			//b.logger.Debugf("transaction validate failed. tx: %v error: %s", transaction, err)
			continue loop
		}

		//b.logger.Debugf("preprocessing transaction: %v", transaction.IERCTransaction)

		switch t := transaction.IERCTransaction.(type) {

		case *protocol.DeployCommand:
			tickSet.Add(t.Tick)

		case *protocol.DeployPoWCommand:
			tickSet.Add(t.Tick)

		case *protocol.MintCommand:
			tickSet.Add(t.Tick)
			balanceSet.Add(balance.NewBalanceKey(t.From, t.Tick))

		case *protocol.MintPoWCommand:
			tickName := t.Tick()
			tickSet.Add(tickName)
			balanceSet.Add(balance.NewBalanceKey(t.From, tickName))
			balanceSet.Add(balance.NewBalanceKey(protocol.ZeroAddress, tickName))

		case *protocol.ModifyCommand:
			tickSet.Add(t.Tick)

		case *protocol.ClaimAirdropCommand:
			tickSet.Add(t.Tick)
			balanceSet.Add(balance.NewBalanceKey(t.From, t.Tick))

		case *protocol.TransferCommand:
			for _, record := range t.Records {
				tickSet.Add(record.Tick)
				balanceSet.Add(balance.NewBalanceKey(record.From, record.Tick))
				balanceSet.Add(balance.NewBalanceKey(record.Recv, record.Tick))
			}

		case *protocol.FreezeSellCommand:
			for _, record := range t.Records {
				tickSet.Add(record.Tick)
				balanceSet.Add(balance.NewBalanceKey(record.Seller, record.Tick))
				signatureSet.Add(record.SellerSign)
			}

		case *protocol.UnfreezeSellCommand:
			for _, record := range t.Records {
				signatureSet.Add(record.Sign)
				unfreezeSignSet.Add(record.Sign)
			}

		case *protocol.ProxyTransferCommand:
			for _, record := range t.Records {
				tickSet.Add(record.Tick)
				balanceSet.Add(balance.NewBalanceKey(record.From, record.Tick))
				balanceSet.Add(balance.NewBalanceKey(record.To, record.Tick))
				signatureSet.Add(record.Sign)
			}

		case *protocol.ConfigStakeCommand:
			for _, record := range t.Details {
				tickSet.Add(record.Tick)
			}

		case *protocol.StakingCommand:
			for _, record := range t.Details {
				tickSet.Add(record.Tick)
				balanceSet.Add(balance.NewBalanceKey(record.Pool, record.Tick))
				balanceSet.Add(balance.NewBalanceKey(record.Staker, record.Tick))
			}

		default:
			transaction.Code = int32(protocol.InvalidProtocolParams)
			transaction.Remark = "invalid operate"
			transaction.IsProcessed = true
			continue loop
		}
	}

	//startAt := time.Now()
	//b.logger.Debugf("start load data. %s", startAt)
	//defer func() {
	//	b.logger.Debugf("load done. duration: %s", time.Since(startAt))
	//}()

	eg, gCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return b.loadTicks(gCtx, aggregate, tickSet.ToSlice())
	})

	eg.Go(func() error {
		return b.loadBalances(gCtx, aggregate, balanceSet.ToSlice())
	})

	eg.Go(func() error {
		return b.loadEventsBySignature(gCtx, aggregate, signatureSet.ToSlice())
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if err := b.loadUnfreezeEventRelatedData(ctx, aggregate, unfreezeSignSet); err != nil {
		return nil, err
	}

	return aggregate, nil
}

func (b *BlockService) loadTicks(ctx context.Context, root *domain.AggregateRoot, names []string) error {
	var queryTicks = make([]string, 0, len(names))
	for _, name := range names {
		_, existed := root.TicksMap[name]
		if existed {
			continue
		}

		queryTicks = append(queryTicks, name)
	}

	if len(queryTicks) == 0 {
		return nil
	}

	for _, tickName := range queryTicks {
		entity, err := b.tickRepo.Load(ctx, tickName)
		if err != nil {
			return err // database error
		}

		if entity == nil {
			continue
		}

		root.TicksMap[entity.GetName()] = entity
	}

	return nil
}

func (b *BlockService) loadBalances(ctx context.Context, root *domain.AggregateRoot, keys []balance.BalanceKey) error {

	var queries = make([]balance.BalanceKey, 0, len(keys))
	for _, key := range keys {
		_, existed := root.BalancesMap[key]
		if existed {
			continue
		}

		queries = append(queries, key)
	}

	if len(queries) == 0 {
		return nil
	}

	for _, key := range queries {
		entity, err := b.balanceRepo.Load(ctx, key)
		if err != nil {
			return err
		}

		if entity == nil {
			continue
		}

		root.BalancesMap[entity.Key()] = entity
	}

	return nil
}

func (b *BlockService) loadEventsBySignature(ctx context.Context, root *domain.AggregateRoot, signs []string) error {
	signatures, err := b.eventRepo.QueryEventBySignature(ctx, signs)
	if err != nil {
		return err
	}

	for _, event := range signatures {
		if event == nil {
			continue
		}

		if e, ok := event.(*domain.IERC20TransferredEvent); ok && e.Data.Sign != "" {
			root.Signatures[e.Data.Sign] = e
		}
	}

	return nil
}

func (b *BlockService) loadUnfreezeEventRelatedData(ctx context.Context, root *domain.AggregateRoot, unfreezeSignSet mapset.Set[string]) error {

	var (
		tickSet    = mapset.NewSet[string]()
		balanceSet = mapset.NewSet[balance.BalanceKey]()
	)

	for sign, e := range root.Signatures {

		if !unfreezeSignSet.Contains(sign) {
			continue
		}

		if e.Data.Operate != protocol.OpFreezeSell {
			continue
		}

		_, existed := root.TicksMap[e.Data.Tick]
		if !existed {
			tickSet.Add(e.Data.Tick)
		}

		key := balance.NewBalanceKey(e.Data.From, e.Data.Tick)
		_, existed = root.BalancesMap[key]
		if !existed {
			balanceSet.Add(key)
		}
	}

	eg, gCtx := errgroup.WithContext(ctx)
	if tickSet.Cardinality() > 0 {
		eg.Go(func() error {
			return b.loadTicks(gCtx, root, tickSet.ToSlice())
		})
	}

	if balanceSet.Cardinality() > 0 {
		eg.Go(func() error {
			return b.loadBalances(gCtx, root, balanceSet.ToSlice())
		})
	}

	return eg.Wait()
}

func (b *BlockService) saveToDBWithTx(ctx context.Context, root *domain.AggregateRoot) error {

	var (
		needUpdateTicks    = make([]tick.Tick, 0, len(root.TicksMap))
		needUpdateBalances = make([]*balance.Balance, 0, len(root.BalancesMap))
		pools              = poolsMapToSlice(root.StakingPools)
	)

	for _, entity := range root.TicksMap {
		if entity.LastUpdatedBlock() < root.Block.Number {
			continue
		}

		needUpdateTicks = append(needUpdateTicks, entity)
	}

	for _, entity := range root.BalancesMap {
		if entity.LastUpdatedBlock < root.Block.Number {
			continue
		}

		needUpdateBalances = append(needUpdateBalances, entity)
	}

	err := b.transactionRepo.TransactionSave(ctx, func(ctxWithTx context.Context) error {
		if err := b.blockRepo.Update(ctxWithTx, root.Block); err != nil {
			return err
		}

		event := &domain.EventsByBlock{BlockNumber: root.Block.Number, Events: root.Events}
		if err := b.eventRepo.Save(ctxWithTx, event); err != nil {
			return err
		}

		if err := b.tickRepo.Save(ctxWithTx, needUpdateTicks...); err != nil {
			return err
		}

		if err := b.balanceRepo.Save(ctxWithTx, needUpdateBalances...); err != nil {
			return err
		}

		if err := b.stakingRepo.Save(ctxWithTx, root.Block.Number, pools...); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return b.transactionRepo.UpdateCache(ctx, func(ctxWithUpdateKind context.Context) error {
		_ = b.tickRepo.Save(ctxWithUpdateKind, needUpdateTicks...)
		_ = b.balanceRepo.Save(ctxWithUpdateKind, needUpdateBalances...)
		_ = b.stakingRepo.Save(ctxWithUpdateKind, root.Block.Number, pools...)
		return nil
	})
}

func poolsMapToSlice(poolsMap map[string]*staking.PoolAggregate) []*staking.PoolAggregate {
	var result = make([]*staking.PoolAggregate, 0, len(poolsMap))
	for _, root := range poolsMap {
		result = append(result, root)
	}

	return result
}
