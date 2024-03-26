package mysqlimpl

import (
	"context"
	"errors"
	"sync"

	"github.com/IErcOrg/IERC_Indexer/internal/domain"
	rctx "github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/context"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/mysql/acl"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/mysql/models"
	"gorm.io/gorm"
)

type eventRepo struct {
	db *gorm.DB

	subscriber map[string]*domain.Stream[domain.EventsByBlock]
	rw         sync.Mutex
}

func NewEventRepository(db *gorm.DB) domain.EventRepository {
	return &eventRepo{
		db:         db,
		subscriber: make(map[string]*domain.Stream[domain.EventsByBlock]),
		rw:         sync.Mutex{},
	}
}

func (repo *eventRepo) GetBlockNumberByLastEvent(ctx context.Context) (uint64, error) {

	var m uint64
	result := repo.db.WithContext(ctx).
		Table((&models.Event{}).TableName()).
		Select("block_number").
		Order("`block_number` DESC").Take(&m)
	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil
		}

		return 0, err
	}

	return m, nil
}

func (repo *eventRepo) QueryEventBySignature(ctx context.Context, signs []string) (map[string]domain.Event, error) {
	var eventsBySign = make(map[string]domain.Event)
	if len(signs) == 0 {
		return eventsBySign, nil
	}

	for _, sign := range signs {

		entity, err := repo.queryEventBySignature(ctx, sign)
		if err != nil {
			return nil, err
		}

		if entity == nil {
			continue
		}

		eventsBySign[sign] = entity
	}

	return eventsBySign, nil
}

func (repo *eventRepo) queryEventBySignature(ctx context.Context, signs string) (domain.Event, error) {
	var m models.Event

	err := repo.db.WithContext(ctx).
		Where("`err_code` = 0 and `sign` = ?", signs).
		Order("`block_number` DESC").
		Take(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return acl.ConvertModelToEvent(&m), nil
}

func (repo *eventRepo) SubscribeEvent(ctx context.Context, startBlock uint64) (*domain.Stream[domain.EventsByBlock], error) {

	stream := domain.NewEventStream[domain.EventsByBlock](100)

	go func() {

		startBlock := startBlock
		for {
			blocks, err := repo.LoadEventsByBlocks(ctx, startBlock, 100)
			if err != nil {
				stream.SendErr(err)
				return
			}

			if len(blocks) == 0 {
				break
			}

			for _, m := range blocks {
				stream.Send(m)
				startBlock = m.BlockNumber
			}

			if blocks[len(blocks)-1].BlockNumber == startBlock {
				break
			}
		}

		repo.rw.Lock()
		repo.subscriber[stream.ID()] = stream
		repo.rw.Unlock()

		select {
		case <-ctx.Done():
			repo.rw.Lock()
			delete(repo.subscriber, stream.ID())
			repo.rw.Unlock()
			stream.Close()
		}
	}()

	return stream, nil
}

func (repo *eventRepo) LoadEventsByBlocks(ctx context.Context, startBlock uint64, limit int) ([]*domain.EventsByBlock, error) {

	var ms []*models.Event
	result := repo.db.WithContext(ctx).
		Table((&models.Event{}).TableName()).
		Where("`block_number` > ?", startBlock).
		Limit(limit).
		Order("`block_number` ASC, `id` ASC").Find(&ms)
	if err := result.Error; err != nil {
		return nil, err
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	lastBlock := ms[len(ms)-1]
	var ms1 []*models.Event
	err := repo.db.WithContext(ctx).Table((&models.Event{}).TableName()).
		Where("`block_number` = ? and id > ?", lastBlock.BlockNumber, lastBlock.ID).
		Order("`id` ASC").Find(&ms1).Error
	if err != nil {
		return nil, err
	}

	ms = append(ms, ms1...)

	var blocksMap = make(map[uint64]*domain.EventsByBlock)
	var blocks []*domain.EventsByBlock
	for _, m := range ms {

		block, existed := blocksMap[m.BlockNumber]
		if !existed {
			block = &domain.EventsByBlock{
				BlockNumber: m.BlockNumber,
				Events:      nil,
			}
			blocksMap[block.BlockNumber] = block
			blocks = append(blocks, block)
		}

		block.Events = append(block.Events, acl.ConvertModelToEvent(m))
	}

	return blocks, nil
}

func (repo *eventRepo) QueryEventsByBlocks(ctx context.Context, startBlock uint64, blockNum int) ([]*domain.EventsByBlock, error) {

	var blockNums []uint64
	err := repo.db.WithContext(ctx).
		Table((&models.Event{}).TableName()).
		Select("`block_number`").
		Where("`block_number` > ?", startBlock).
		Group("`block_number`").
		Limit(blockNum).
		Order("`block_number`").
		Find(&blockNums).
		Error
	if err != nil {
		return nil, err
	}

	if len(blockNums) == 0 {
		return nil, nil
	}

	var ms []*models.Event
	result := repo.db.WithContext(ctx).
		Table((&models.Event{}).TableName()).
		Where("`block_number` in ?", blockNums).
		Order("`block_number` ASC, `id` ASC").
		Find(&ms)
	if err := result.Error; err != nil {
		return nil, err
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	var blocksMap = make(map[uint64]*domain.EventsByBlock)
	var blocks []*domain.EventsByBlock
	for _, m := range ms {

		block, existed := blocksMap[m.BlockNumber]
		if !existed {
			block = &domain.EventsByBlock{
				BlockNumber: m.BlockNumber,
				Events:      nil,
			}
			blocksMap[block.BlockNumber] = block
			blocks = append(blocks, block)
		}

		block.Events = append(block.Events, acl.ConvertModelToEvent(m))
	}

	return blocks, nil
}

func (repo *eventRepo) QueryEventsByHash(ctx context.Context, hash string) ([]domain.Event, error) {

	var ms []*models.Event
	err := repo.db.WithContext(ctx).
		Table((&models.Event{}).TableName()).
		Where("`tx_hash` = ?", hash).
		Order("`id` ASC").
		Find(&ms).
		Error
	if err != nil {
		return nil, err
	}

	var events = make([]domain.Event, 0, len(ms))
	for _, m := range ms {
		events = append(events, acl.ConvertModelToEvent(m))
	}

	return events, nil
}

func (repo *eventRepo) Save(ctx context.Context, event *domain.EventsByBlock) error {

	if len(event.Events) == 0 {
		return nil
	}

	dbWithTx := rctx.TransactionDBFromContext(ctx)
	if dbWithTx == nil {
		panic("missing db instance")
	}

	var ms []*models.Event
	for _, entity := range event.Events {
		ms = append(ms, acl.ConvertEventToModel(entity))
	}

	if err := dbWithTx.CreateInBatches(ms, 1000).Error; err != nil {
		return err
	}

	return repo.publishEvents(ctx, event)
}

func (repo *eventRepo) publishEvents(ctx context.Context, event *domain.EventsByBlock) error {
	if len(repo.subscriber) == 0 {
		return nil
	}

	repo.rw.Lock()
	defer repo.rw.Unlock()
	for _, stream := range repo.subscriber {
		select {
		case stream.Input() <- event:
		default:
		}
	}

	return nil
}
