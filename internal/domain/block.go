package domain

import (
	"fmt"
	"strconv"
	"time"

	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
	"github.com/shopspring/decimal"
)

type BlockHeader struct {
	Number     uint64
	Hash       string
	ParentHash string
}

func (b *BlockHeader) String() string {
	if b == nil {
		return "0"
	} else {
		return strconv.FormatUint(b.Number, 10)
	}
}

type BlockHandleStatus struct {
	LatestBlock      *BlockHeader
	LastIndexedBlock *BlockHeader
	LastSyncBlock    *BlockHeader
}

func (b *BlockHandleStatus) String() string {
	if b == nil {
		return "nil"
	}

	return fmt.Sprintf("latestBlock: %s, indexedBlock: %s, syncBlock: %s", b.LatestBlock, b.LastIndexedBlock, b.LastSyncBlock)
}

type Block struct {
	Number           uint64
	ParentHash       string
	Hash             string
	TransactionCount int
	Transactions     []*Transaction
	IsProcessed      bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (b *Block) Header() *BlockHeader {
	return &BlockHeader{
		Number:     b.Number,
		Hash:       b.Hash,
		ParentHash: b.ParentHash,
	}
}

type Transaction struct {
	BlockNumber   uint64
	PositionInTxs int64
	Hash          string
	From          string
	To            string
	TxData        string
	TxValue       decimal.Decimal
	Gas           decimal.Decimal
	GasPrice      decimal.Decimal
	Nonce         uint64

	IsProcessed bool
	Code        int32
	Remark      string
	CreatedAt   time.Time
	UpdatedAt   time.Time

	IERCTransaction protocol.IERCTransaction
}
