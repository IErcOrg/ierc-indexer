package ethereum

import (
	"context"
	"errors"
	"math/big"
	"time"

	"github.com/IErcOrg/IERC_Indexer/internal/conf"
	"github.com/IErcOrg/IERC_Indexer/internal/domain"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol/parser"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

type EthereumFetcher struct {
	clis   []*ethclient.Client
	parser parser.Parser
	logger *log.Helper
}

func NewEthereumFetcher(conf *conf.Config, parser parser.Parser, logger log.Logger) (domain.BlockFetcher, error) {
	data := conf.Bootstrap.Data

	if len(data.Ethereum.Endpoints) == 0 {
		return nil, errors.New("missing ethereum rpc endpoints")
	}

	var clis []*ethclient.Client
	for _, endpoint := range data.Ethereum.Endpoints {
		c, err := rpc.DialOptions(context.Background(), endpoint)
		if err != nil {
			return nil, err
		}

		clis = append(clis, ethclient.NewClient(c))
	}

	return &EthereumFetcher{
		clis:   clis,
		parser: parser,
		logger: log.NewHelper(log.With(logger, "module", "fetcher")),
	}, nil
}

func (e *EthereumFetcher) GetBlockNumber(ctx context.Context) (uint64, error) {
	return e.clis[0].BlockNumber(ctx)
}

func (e *EthereumFetcher) GetBlockHeaderByNumber(ctx context.Context, blockNumber uint64) (*domain.BlockHeader, error) {
	var params *big.Int
	if blockNumber != 0 {
		params = new(big.Int).SetUint64(blockNumber)
	}
	header, err := e.clis[0].HeaderByNumber(ctx, params)
	if err != nil {
		return nil, err
	}

	return &domain.BlockHeader{
		Number:     header.Number.Uint64(),
		Hash:       header.Hash().String(),
		ParentHash: header.ParentHash.String(),
	}, nil
}

func (e *EthereumFetcher) GetBlockByNumber(ctx context.Context, targetBlock uint64) (*domain.Block, error) {
	var block *types.Block

	for idx, cli := range e.clis {
		cli := cli
		newBlock, err := cli.BlockByNumber(ctx, new(big.Int).SetUint64(targetBlock))
		if err != nil {
			return nil, err
		}

		if block != nil && (newBlock.Hash() != block.Hash() || newBlock.Transactions().Len() != block.Transactions().Len()) {
			e.logger.Warnf(
				"block hash not match. client_idx: %d, last_number: %d, last_hash: %v, tx_count: %d, current_number: %d, current_hash: %v, tx_count: %d",
				idx, block.NumberU64(), block.Hash(), block.Transactions().Len(), newBlock.NumberU64(), newBlock.Hash(), newBlock.Transactions().Len(),
			)
			return nil, errors.New("block inconsistent")
		}

		block = newBlock
	}

	return e.parseBlock(block)
}

func GetTxSender(tx *types.Transaction) (common.Address, error) {
	var signer types.Signer
	switch {
	case tx.Type() == types.AccessListTxType:
		signer = types.NewEIP2930Signer(tx.ChainId())
	case tx.Type() == types.DynamicFeeTxType:
		signer = types.NewLondonSigner(tx.ChainId())
	default:
		signer = types.NewEIP155Signer(tx.ChainId())
	}
	sender, err := types.Sender(signer, tx)
	return sender, err
}

func (e *EthereumFetcher) parseBlock(block *types.Block) (*domain.Block, error) {

	var transactions []*domain.Transaction
	for position, tx := range block.Transactions() {

		err := e.parser.CheckFormat(tx.Data())
		if err != nil {
			continue
		}

		from, err := GetTxSender(tx)
		if err != nil {
			return nil, err
		}

		to := protocol.ZeroAddress
		if tx.To() != nil {
			to = tx.To().String()
		}

		transactions = append(transactions, &domain.Transaction{
			BlockNumber:     block.NumberU64(),
			PositionInTxs:   int64(position),
			Hash:            tx.Hash().String(),
			From:            from.String(),
			To:              to,
			TxData:          string(tx.Data()),
			TxValue:         decimal.NewFromBigInt(tx.Value(), 0),
			Gas:             decimal.NewFromBigInt(new(big.Int).SetUint64(tx.Gas()), 0),
			GasPrice:        decimal.NewFromBigInt(tx.GasPrice(), 0),
			Nonce:           tx.Nonce(),
			IsProcessed:     false,
			Code:            0,
			Remark:          "",
			CreatedAt:       time.Unix(int64(block.Time()), 0),
			UpdatedAt:       time.Unix(int64(block.Time()), 0),
			IERCTransaction: nil,
		})
	}

	return &domain.Block{
		Number:           block.NumberU64(),
		ParentHash:       block.ParentHash().String(),
		Hash:             block.Hash().String(),
		TransactionCount: len(transactions),
		Transactions:     transactions,
		IsProcessed:      len(transactions) == 0,
	}, nil
}
