package protocol

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

func TestName(t *testing.T) {
	fmt.Print(common.IsHexAddress("1x1111cccc5dfa575bb183077a1c0f525cf7b50d48"))
}

func TestProtocol(t *testing.T) {
	suite.Run(t, new(TestProtocolSuite))
}

type TestProtocolSuite struct {
	suite.Suite
}

func (s *TestProtocolSuite) SetupSuite() {

}

func (s *TestProtocolSuite) TestStakingCommand() {
	stakingCommand := &StakingCommand{
		IERCTransactionBase: IERCTransactionBase{
			BlockNumber:        1000,
			TxHash:             "",
			TxValue:            decimal.Zero,
			PositionInBlockTxs: 0,
			From:               "",
			To:                 "",
			Protocol:           ProtocolIERC20,
			Operate:            OpStaking,
		},
		Pool:      "0x00",
		PoolSubID: 1,
		Details: []*StakingDetail{{
			Staker: "0x000012",
			Tick:   "ethi",
			Amount: decimal.NewFromInt(10000),
		}},
	}

	unstakingCommand := &StakingCommand{
		IERCTransactionBase: IERCTransactionBase{
			BlockNumber:        1000,
			TxHash:             "",
			TxValue:            decimal.Zero,
			PositionInBlockTxs: 0,
			From:               "",
			To:                 "",
			Protocol:           ProtocolIERC20,
			Operate:            OpUnStaking,
		},
		Pool:      "0x00",
		PoolSubID: 1,
		Details: []*StakingDetail{{
			Staker: "0x000012",
			Tick:   "ethi",
			Amount: decimal.NewFromInt(10000),
		}},
	}

	fmt.Printf("--1: %s\n", stakingCommand.String())
	fmt.Printf("--2: %s\n", unstakingCommand.String())

	for _, a := range append([]any(nil), stakingCommand, unstakingCommand) {
		switch aa := a.(type) {
		case *StakingCommand:
			fmt.Printf("%s\n", aa.String())
		default:
			//fmt.Printf("%v\n", aa)
		}
	}
}
