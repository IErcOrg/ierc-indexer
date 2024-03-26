package domain

import (
	"context"

	"github.com/google/uuid"
)

type BlockFetcher interface {
	GetBlockNumber(ctx context.Context) (uint64, error)
	GetBlockHeaderByNumber(ctx context.Context, blockNumber uint64) (*BlockHeader, error)
	GetBlockByNumber(ctx context.Context, targetBlock uint64) (*Block, error)
}

type BlockRepository interface {
	GetLastIndexedBlock(ctx context.Context) (*BlockHeader, error)
	GetLastHandleBlock(ctx context.Context) (*BlockHeader, error)

	GetPendingBlocksWithTransactionsByNumber(ctx context.Context, number uint64, bulkSize int) ([]*Block, error)

	QueryLastProcessedBlock(ctx context.Context, blockNumber uint64) (*BlockHeader, error)
	QueryTransactionByHash(ctx context.Context, hash string) (*Transaction, error)

	BulkSaveBlock(ctx context.Context, blocks []*Block) error
	Update(ctx context.Context, block *Block) error
}

type Stream[T any] struct {
	id     string
	dataCh chan *T
	errCh  chan error
}

func NewEventStream[T any](size int) *Stream[T] {
	return &Stream[T]{
		id:     uuid.NewString(),
		dataCh: make(chan *T, size),
		errCh:  make(chan error),
	}
}

func (s *Stream[T]) ID() string {
	return s.id
}

func (s *Stream[T]) Input() chan<- *T {
	return s.dataCh
}

func (s *Stream[T]) Send(data *T) {
	s.dataCh <- data
}

func (s *Stream[T]) Next() <-chan *T {
	return s.dataCh
}

func (s *Stream[T]) SendErr(err error) {
	select {
	case s.errCh <- err:
		close(s.errCh)
	}
}

func (s *Stream[T]) Err() <-chan error {
	return s.errCh
}

func (s *Stream[T]) Close() {
	close(s.dataCh)
	close(s.errCh)
}

type EventRepository interface {
	Save(ctx context.Context, event *EventsByBlock) error

	GetBlockNumberByLastEvent(ctx context.Context) (uint64, error)
	QueryEventBySignature(ctx context.Context, signs []string) (map[string]Event, error)
	SubscribeEvent(ctx context.Context, startBlock uint64) (*Stream[EventsByBlock], error)
	LoadEventsByBlocks(ctx context.Context, startBlock uint64, limit int) ([]*EventsByBlock, error)
	QueryEventsByBlocks(ctx context.Context, startBlock uint64, blockNum int) ([]*EventsByBlock, error)
	QueryEventsByHash(ctx context.Context, hash string) ([]Event, error)
}

type TransactionRepository interface {
	TransactionSave(ctx context.Context, fn func(ctx context.Context) error) error
	UpdateCache(ctx context.Context, fn func(ctx context.Context) error) error
}
