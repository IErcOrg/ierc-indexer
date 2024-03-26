package acl

import (
	"github.com/IErcOrg/IERC_Indexer/internal/domain"
	"github.com/IErcOrg/IERC_Indexer/internal/infrastructure/repository/mysql/models"
)

func ConvertBlockEntityToModel(block *domain.Block) *models.Block {
	return &models.Block{
		ID:               0,
		Number:           block.Number,
		Hash:             block.Hash,
		ParentHash:       block.ParentHash,
		TransactionCount: block.TransactionCount,
		IsProcessed:      block.IsProcessed,
		CreatedAt:        block.CreatedAt,
		UpdatedAt:        block.UpdatedAt,
	}
}

func ConvertTransactionEntityToModel(tx *domain.Transaction) *models.Transaction {
	return &models.Transaction{
		ID:            0,
		BlockNumber:   tx.BlockNumber,
		PositionInTxs: tx.PositionInTxs,
		Hash:          tx.Hash,
		From:          tx.From,
		To:            tx.To,
		Value:         tx.TxValue,
		Gas:           tx.Gas,
		GasPrice:      tx.GasPrice,
		Data:          tx.TxData,
		Nonce:         tx.Nonce,
		IsProcessed:   tx.IsProcessed,
		Code:          tx.Code,
		Remark:        tx.Remark,
		CreatedAt:     tx.CreatedAt,
		UpdatedAt:     tx.UpdatedAt,
	}
}

// entity => model
func BulkConvertTransactionEntityToModel(txs []*domain.Transaction) []*models.Transaction {
	var transactions = make([]*models.Transaction, 0, len(txs))
	for _, tx := range txs {
		transactions = append(transactions, ConvertTransactionEntityToModel(tx))
	}

	return transactions
}

// model => entity
func ConvertTransactionModelToEntity(tx *models.Transaction) *domain.Transaction {
	return &domain.Transaction{
		BlockNumber:     tx.BlockNumber,
		PositionInTxs:   tx.PositionInTxs,
		Hash:            tx.Hash,
		From:            tx.From,
		To:              tx.To,
		TxData:          tx.Data,
		TxValue:         tx.Value,
		Gas:             tx.Gas,
		GasPrice:        tx.GasPrice,
		Nonce:           tx.Nonce,
		IsProcessed:     tx.IsProcessed,
		Code:            tx.Code,
		Remark:          tx.Remark,
		CreatedAt:       tx.CreatedAt,
		UpdatedAt:       tx.UpdatedAt,
		IERCTransaction: nil,
	}
}
