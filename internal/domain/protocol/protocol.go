package protocol

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type IERCTransaction interface {
	String() string
	Validate() error
}

var (
	_ IERCTransaction = (*DeployCommand)(nil)
	_ IERCTransaction = (*DeployPoWCommand)(nil)
	_ IERCTransaction = (*MintCommand)(nil)
	_ IERCTransaction = (*MintPoWCommand)(nil)
	_ IERCTransaction = (*TransferCommand)(nil)
	_ IERCTransaction = (*FreezeSellCommand)(nil)
	_ IERCTransaction = (*ProxyTransferCommand)(nil)
	_ IERCTransaction = (*ConfigStakeCommand)(nil)
	_ IERCTransaction = (*StakingCommand)(nil)
)

type IERCTransactionBase struct {
	BlockNumber        uint64          `json:"-"`
	TxHash             string          `json:"-"`
	TxValue            decimal.Decimal `json:"-"`
	PositionInBlockTxs int64           `json:"-"`
	From               string          `json:"-"`
	To                 string          `json:"-"`
	Gas                decimal.Decimal `json:"-"`
	GasPrice           decimal.Decimal `json:"-"`
	EventAt            time.Time       `json:"-"`

	Protocol Protocol `json:"p"`
	Operate  Operate  `json:"op"`
}

func (protocol *IERCTransactionBase) String() string {
	return fmt.Sprintf(
		"block: %d, hash: %s, from: %s, to: %s, protocol: %s, operate: %s",
		protocol.BlockNumber, protocol.TxHash, protocol.From, protocol.To, protocol.Protocol, protocol.Operate,
	)
}

func (protocol *IERCTransactionBase) Validate() error {

	switch protocol.Operate {

	//case OpDeploy, OpMint, OpTransfer, OpStakeConfig, OpStaking, OpUnStaking, OpProxyUnStaking:
	default:
		// check: to == 0x0000000000000000000000000000000000000000
		if protocol.To != ZeroAddress {
			return NewProtocolError(InvalidProtocolParams, "invalid to address. must be zero address")
		}

	case OpFreezeSell:
		if protocol.To != PlatformAddress {
			return NewProtocolError(InvalidProtocolParams, "invalid to address. must be platform address")
		}

	case OpUnfreezeSell, OpProxyTransfer:
		if protocol.From != PlatformAddress {
			return NewProtocolError(InvalidProtocolParams, "invalid from address. must be platform address")
		}
	}

	return nil
}
