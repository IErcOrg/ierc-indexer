package service

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/IErcOrg/IERC_Indexer/internal/conf"
	"github.com/IErcOrg/IERC_Indexer/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

func TestBlockHandler(t *testing.T) {
	suite.Run(t, new(TestBlockHandlerSuite))
}

type TestBlockHandlerSuite struct {
	suite.Suite
	handler *BlockService
}

func (s *TestBlockHandlerSuite) SetupSuite() {
	var data = conf.Config{
		Config: nil,
		Bootstrap: &conf.Bootstrap{
			Server: nil,
			Data:   nil,
			Runtime: &conf.Runtime{
				EnableSync:        false,
				SyncStartBlock:    0,
				SyncThreadsNum:    0,
				EnableHandle:      false,
				HandleEndBlock:    0,
				HandleQueueSize:   0,
				InvalidTxHashPath: "",
				FeeStartBlock:     0,
			},
		},
		InvalidTxHash: nil,
	}

	_ = data
	var (
	//aggrRepo    = mock.NewMockAggregateRepository()
	//stakingRepo = mock.NewMockStakingRepository()
	)

	//h, err := NewBlockService(&data, log.DefaultLogger, aggrRepo, stakingRepo)
	//s.NoError(err)
	//s.handler = h
}

func (s *TestBlockHandlerSuite) TestStaking() {

	_ = &domain.Block{
		Number:           0,
		ParentHash:       "",
		Hash:             "",
		TransactionCount: 0,
		Transactions: []*domain.Transaction{
			{
				BlockNumber:     0,
				PositionInTxs:   0,
				Hash:            "",
				From:            "0x000",
				To:              "0x111",
				TxData:          ``,
				TxValue:         decimal.Decimal{},
				GasPrice:        decimal.Decimal{},
				Nonce:           0,
				IERCTransaction: nil,
				IsProcessed:     false,
				Code:            0,
				Remark:          "",
				CreatedAt:       time.Time{},
				UpdatedAt:       time.Time{},
			},
		},
		IsProcessed: false,
		CreatedAt:   time.Time{},
		UpdatedAt:   time.Time{},
	}

}

func (s *TestBlockHandlerSuite) TestDeferReturn() {
	deferReturn()
}

func deferReturn() (err error) {
	defer func() {
		fmt.Print(err)
	}()

	aa, err := ss()
	_ = aa

	return err
}

func ss() (any, error) {
	return nil, errors.New("----------")
}
