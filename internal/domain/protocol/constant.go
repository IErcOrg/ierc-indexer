package protocol

import (
	"regexp"
	"strings"

	"github.com/IErcOrg/IERC_Indexer/pkg/utils"
	"github.com/shopspring/decimal"
)

const (
	ZeroAddress = "0x0000000000000000000000000000000000000000"

	ProtocolHeader = `data:application/json,`

	TickETHI = "ethi"

	SignatureTitle = "ierc-20 one approve"

	TickMaxLength = 64
)

var (
	TickMaxSupplyLimit = decimal.RequireFromString("9999999999999999999999999999999")
	ServiceFee         = decimal.RequireFromString("1.02")
	WorkCRegexp        = regexp.MustCompile(`^0x[0-9a-f]{1,64}$`)
)

func init() {

	if !utils.IsHexAddressWith0xPrefix(ZeroAddress) ||
		!utils.IsHexAddressWith0xPrefix(PlatformAddress) ||
		strings.ToLower(ZeroAddress) != ZeroAddress ||
		strings.ToLower(PlatformAddress) != PlatformAddress {
		panic("constant check error")
	}
}

type Protocol string

const (
	ProtocolTERC20  Protocol = "terc-20"
	ProtocolIERC20  Protocol = "ierc-20"
	ProtocolIERCPoW Protocol = "ierc-pow"
)

type Operate string

const (
	OpDeploy         = "deploy"
	OpMint           = "mint"
	OpTransfer       = "transfer"
	OpFreezeSell     = "freeze_sell"
	OpUnfreezeSell   = "unfreeze_sell"
	OpRefund         = "refund"
	OpProxyTransfer  = "proxy_transfer"
	OpStakeConfig    = "stake_config"
	OpStaking        = "stake"
	OpUnStaking      = "unstake"
	OpProxyUnStaking = "proxy_unstake"

	OpPoWModify       = "modify"
	OpPoWClaimAirdrop = "airdrop_claim"
)
