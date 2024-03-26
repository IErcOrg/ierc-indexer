package mysqlimpl

import (
	"context"
	"errors"

	"github.com/IErcOrg/IERC_Indexer/internal/domain"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol/parser"
	rctx "github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/context"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/mysql/acl"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/mysql/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type blockMySQLRepo struct {
	db     *gorm.DB
	parser parser.Parser
}

func NewBlockRepo(db *gorm.DB, parser parser.Parser) domain.BlockRepository {
	return &blockMySQLRepo{
		db:     db,
		parser: parser,
	}
}

func (repo *blockMySQLRepo) GetLastIndexedBlock(ctx context.Context) (*domain.BlockHeader, error) {

	var block models.Block
	err := repo.db.WithContext(ctx).
		Table(block.TableName()).
		Order("block_number DESC").
		Take(&block).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &domain.BlockHeader{
		Number:     block.Number,
		Hash:       block.Hash,
		ParentHash: block.ParentHash,
	}, nil
}

func (repo *blockMySQLRepo) GetLastHandleBlock(ctx context.Context) (*domain.BlockHeader, error) {

	var block models.Block
	err := repo.db.WithContext(ctx).
		Table(block.TableName()).
		Where("tx_count > 0 and is_processed = 1").
		Order("block_number DESC").
		Take(&block).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &domain.BlockHeader{
		Number:     block.Number,
		Hash:       block.Hash,
		ParentHash: block.ParentHash,
	}, nil
}

func (repo *blockMySQLRepo) QueryLastProcessedBlock(ctx context.Context, blockNumber uint64) (*domain.BlockHeader, error) {

	var block models.Block
	result := repo.db.WithContext(ctx).
		Table(block.TableName()).
		Where("block_number > ? and tx_count > 0 and is_processed = 0", blockNumber).
		Order("block_number ASC").
		Take(&block)
	if err := result.Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if result.RowsAffected > 0 {
		return &domain.BlockHeader{
			Number:     block.Number - 1,
			Hash:       block.ParentHash,
			ParentHash: "",
		}, nil
	}

	result = repo.db.WithContext(ctx).
		Table(block.TableName()).
		Where("block_number > ? and is_processed = 1", blockNumber).
		Order("block_number DESC").
		Take(&block)
	if err := result.Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &domain.BlockHeader{
		Number:     block.Number,
		Hash:       block.Hash,
		ParentHash: block.ParentHash,
	}, nil
}

func (repo *blockMySQLRepo) GetPendingBlocksWithTransactionsByNumber(ctx context.Context, number uint64, bulkSize int) ([]*domain.Block, error) {
	var (
		block  models.Block
		blocks = make([]*models.Block, 0, bulkSize)
	)

	err := repo.db.WithContext(ctx).
		Table(block.TableName()).
		Where("block_number > ? and tx_count > 0 and is_processed = 0", number).
		Order("block_number ASC").
		Limit(bulkSize).
		Find(&blocks).Error

	if err != nil {
		return nil, err
	}

	var result = make([]*domain.Block, 0, len(blocks))
	for _, block := range blocks {

		b := domain.Block{
			Number:           block.Number,
			ParentHash:       block.ParentHash,
			Hash:             block.Hash,
			TransactionCount: block.TransactionCount,
			Transactions:     nil,
			IsProcessed:      block.IsProcessed,
			CreatedAt:        block.CreatedAt,
			UpdatedAt:        block.UpdatedAt,
		}

		if b.TransactionCount != 0 {
			txs, err := repo.QueryTransactions(ctx, block.Number)
			if err != nil {
				return nil, err
			}

			b.Transactions = txs
		}

		result = append(result, &b)
	}

	return result, nil
}

func (repo *blockMySQLRepo) QueryTransactions(ctx context.Context, blockNumber uint64) ([]*domain.Transaction, error) {
	var tx models.Transaction
	var txs []*models.Transaction

	err := repo.db.WithContext(ctx).
		Table(tx.TableName()).
		Where("block_number = ?", blockNumber).
		Order("block_number,position ASC").
		Find(&txs).Error
	if err != nil {
		return nil, err
	}

	var transactions = make([]*domain.Transaction, 0, len(txs))
	for _, tx := range txs {
		transaction := acl.ConvertTransactionModelToEntity(tx)

		tx, err := repo.parser.Parse(transaction)
		if err != nil {
			var (
				code    int32
				message string
			)
			var pErr *protocol.ProtocolError
			if errors.As(err, &pErr) {
				code = pErr.Code()
				message = pErr.Message()
			} else {
				code = int32(protocol.UnknownError)
				message = err.Error()
			}

			transaction.IsProcessed = true
			transaction.Code = code
			transaction.Remark = message
		} else {
			transaction.IERCTransaction = tx
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (repo *blockMySQLRepo) QueryTransactionByHash(ctx context.Context, hash string) (*domain.Transaction, error) {
	var m models.Transaction

	err := repo.db.WithContext(ctx).Table(m.TableName()).Where("hash = ?", hash).Take(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("not found")
		}

		return nil, err
	}

	transaction := acl.ConvertTransactionModelToEntity(&m)

	tx, err := repo.parser.Parse(transaction)
	if err != nil {
		var (
			code    int32
			message string
		)
		var pErr *protocol.ProtocolError
		if errors.As(err, &pErr) {
			code = pErr.Code()
			message = pErr.Message()
		} else {
			code = int32(protocol.UnknownError)
			message = err.Error()
		}

		transaction.IsProcessed = true
		transaction.Code = code
		transaction.Remark = message
	} else {
		transaction.IERCTransaction = tx
	}

	return transaction, nil
}

func (repo *blockMySQLRepo) BulkSaveBlock(ctx context.Context, blocks []*domain.Block) error {

	var (
		bs           = make([]*models.Block, 0, len(blocks))
		transactions []*models.Transaction
	)

	for _, block := range blocks {
		bs = append(bs, acl.ConvertBlockEntityToModel(block))

		for _, transaction := range block.Transactions {
			transactions = append(transactions, &models.Transaction{
				ID:            0,
				BlockNumber:   block.Number,
				PositionInTxs: transaction.PositionInTxs,
				Hash:          transaction.Hash,
				From:          transaction.From,
				To:            transaction.To,
				Value:         transaction.TxValue,
				Gas:           transaction.Gas,
				GasPrice:      transaction.GasPrice,
				Data:          transaction.TxData,
				Nonce:         transaction.Nonce,
				IsProcessed:   transaction.IsProcessed,
				Code:          transaction.Code,
				Remark:        transaction.Remark,
				CreatedAt:     transaction.CreatedAt,
				UpdatedAt:     transaction.UpdatedAt,
			})
		}
	}

	return repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {

		err := tx.CreateInBatches(bs, 1000).Error
		if err != nil {
			return err
		}

		return tx.CreateInBatches(transactions, 1000).Error
	})
}

func (repo *blockMySQLRepo) Update(ctx context.Context, block *domain.Block) error {

	dbWithTx := rctx.TransactionDBFromContext(ctx)
	if dbWithTx == nil {
		panic("missing db instance")
	}

	err := dbWithTx.Table((&models.Block{}).TableName()).
		Where("block_number = ?", block.Number).
		Update("is_processed", true).
		Error
	if err != nil {
		return err
	}

	var transactions = acl.BulkConvertTransactionEntityToModel(block.Transactions)
	return dbWithTx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: `block_number`}, {Name: `position`}},
		DoUpdates: clause.AssignmentColumns([]string{`is_processed`, `code`, `remark`, `updated_at`}),
	}).CreateInBatches(transactions, 1000).Error
}
