package parser

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/IErcOrg/IERC_Indexer/internal/domain"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

// IERC-20
type (
	IERC20 struct {
		Protocol string                `json:"p"`
		Op       string                `json:"op"`
		Tick     string                `json:"tick"`
		Amt      string                `json:"amt"`
		Workc    string                `json:"workc"`
		Nonce    interface{}           `json:"nonce"`
		Max      string                `json:"max"`
		Lim      string                `json:"lim"`
		Wlim     string                `json:"wlim"`
		Dec      string                `json:"dec"`
		To       []IERC20Transfer      `json:"to"`
		Freeze   []IERC20Freeze        `json:"freeze"`
		Unfreeze []IERC20Unfreeze      `json:"unfreeze"`
		Proxy    []IERC20ProxyTransfer `json:"proxy"`
	}

	IERC20Transfer struct {
		Recv string      `json:"recv"`
		Amt  interface{} `json:"amt"`
	}

	IERC20Freeze struct {
		Tick     string      `json:"tick"`
		Platform string      `json:"platform"`
		Seller   string      `json:"seller"`
		Amt      interface{} `json:"amt"`
		Value    string      `json:"value"`
		GasPrice string      `json:"gasPrice"`
		Sign     string      `json:"sign"`
		Nonce    string      `json:"nonce"`
	}

	IERC20Unfreeze struct {
		TxHash              string `json:"txHash"`
		PositionInIERC20Txs int32  `json:"position"`
		Sign                string `json:"sign"`
		Msg                 string `json:"msg"`
	}

	IERC20ProxyTransfer struct {
		Tick  string      `json:"tick"`
		From  string      `json:"from"`
		To    string      `json:"to"`
		Amt   interface{} `json:"amt"`
		Value string      `json:"value"`
		Sign  string      `json:"sign"`
		Nonce string      `json:"nonce"`
	}

	// config staking pool
	ConfigStakingTickDetail struct {
		Tick                 string          `json:"tick"`
		RewardsRatioPerBlock decimal.Decimal `json:"ratio"`
		MaxAmount            decimal.Decimal `json:"max_amt"`
	}
	ConfigStaking struct {
		Pool      string                     `json:"pool"`
		PoolSubID Uint64                     `json:"id"`
		Name      string                     `json:"name"`
		Owner     string                     `json:"owner"`
		Details   []*ConfigStakingTickDetail `json:"details"`
		StopBlock Uint64                     `json:"stop_block,omitempty"`
		MaxAmount decimal.Decimal            `json:"max_amount"`
	}

	// stake & unstake & proxy_unstake
	StakingDetail struct {
		Staker string          `json:"staker"`
		Tick   string          `json:"tick"`
		Amount decimal.Decimal `json:"amt"`
	}
	Staking struct {
		Pool      string           `json:"pool"`
		PoolSubID Uint64           `json:"id"`
		Details   []*StakingDetail `json:"details"`
	}
)

type IERC20Parser struct {
	header       string
	headerLength int
	ethi         string
}

func NewIERC20Parser(header, tick string) *IERC20Parser {
	return &IERC20Parser{
		header:       header,
		headerLength: len(header),
		ethi:         tick,
	}
}

func (parser *IERC20Parser) CheckFormat(data []byte) error {
	return nil
}

func (parser *IERC20Parser) Parse(tx *domain.Transaction) (protocol.IERCTransaction, error) {
	var data = []byte(tx.TxData[parser.headerLength:])
	var ierc20 IERC20
	if err := json.Unmarshal(data, &ierc20); err != nil {
		return nil, protocol.NewProtocolError(protocol.InvalidProtocolFormat, "invalid protocol format")
	}

	base := protocol.IERCTransactionBase{
		BlockNumber:        tx.BlockNumber,
		TxHash:             tx.Hash,
		TxValue:            tx.TxValue,
		PositionInBlockTxs: tx.PositionInTxs,
		From:               strings.ToLower(tx.From),
		To:                 strings.ToLower(tx.To),
		Gas:                tx.Gas,
		GasPrice:           tx.GasPrice,
		EventAt:            tx.CreatedAt,
		Protocol:           protocol.Protocol(ierc20.Protocol),
		Operate:            protocol.Operate(ierc20.Op),
	}

	if err := base.Validate(); err != nil {
		return nil, err
	}

	switch base.Operate {
	case protocol.OpDeploy:
		return parser.parseDeploy(base, &ierc20)

	case protocol.OpMint:
		return parser.parseMint(base, &ierc20)

	case protocol.OpTransfer:
		return parser.parseTransfer(base, &ierc20)

	case protocol.OpFreezeSell:
		return parser.parseFreezeSell(base, &ierc20)

	case protocol.OpUnfreezeSell:
		return parser.parseUnfreezeSell(base, &ierc20)

	case protocol.OpProxyTransfer:
		return parser.parseProxyTransfer(base, &ierc20)

	case protocol.OpStakeConfig:
		return parser.parseConfigStaking(base, data)

	case protocol.OpStaking, protocol.OpUnStaking:
		return parser.parseStakingOrUnStaking(base, data)

	case protocol.OpProxyUnStaking:
		return parser.parseProxyUnStaking(base, data)

	case protocol.OpRefund:
		log.Errorf("refund operate. tx_hash: %s", base.TxHash)
		return nil, protocol.NewProtocolError(protocol.UnknownProtocolOperate, "unknown operate")

	default:
		return nil, protocol.NewProtocolError(protocol.UnknownProtocolOperate, "unknown operate")
	}
}

func (parser *IERC20Parser) parseNonce(tick string, value interface{}) (decimal.Decimal, error) {
	if tick == parser.ethi {

		str := fmt.Sprintf("%s", value)
		if strings.HasPrefix(str, "0") {
			return decimal.Zero, protocol.NewProtocolError(protocol.InvalidProtocolParams, fmt.Sprintf("invalid nonce(%v)", value))
		}

		dec, err := decimal.NewFromString(str)
		if err != nil || dec.LessThan(decimal.Zero) {
			return decimal.Zero, protocol.NewProtocolError(protocol.InvalidProtocolParams, fmt.Sprintf("invalid nonce(%s)", value))
		}

		return dec, nil

	} else {

		var dec decimal.Decimal
		switch v := value.(type) {
		case float64, float32:
			dec = decimal.NewFromFloat(v.(float64))
			if dec.Exponent() < 0 {
				return decimal.Zero, protocol.NewProtocolError(protocol.InvalidProtocolParams, fmt.Sprintf("invalid nonce(%v)", value))
			}

		case string:
			if strings.HasPrefix(v, "0") {
				return decimal.Zero, protocol.NewProtocolError(protocol.InvalidProtocolParams, fmt.Sprintf("invalid nonce(%v)", value))
			}

			num := new(big.Int)
			_, success := num.SetString(v, 10)
			if !success {
				return decimal.Zero, protocol.NewProtocolError(protocol.InvalidProtocolParams, fmt.Sprintf("invalid nonce(%s)", value))
			}

			dec = decimal.NewFromBigInt(num, 0)

		default:
			num := new(big.Int)
			_, success := num.SetString(fmt.Sprintf("%v", value), 10)
			if !success {
				return decimal.Zero, protocol.NewProtocolError(protocol.InvalidProtocolParams, fmt.Sprintf("invalid nonce(%s)", value))
			}

			dec = decimal.NewFromBigInt(num, 0)
		}

		if dec.LessThan(decimal.Zero) {
			return decimal.Zero, protocol.NewProtocolError(protocol.InvalidProtocolParams, fmt.Sprintf("invalid nonce(%s)", value))
		}

		return dec, nil
	}
}

func (parser *IERC20Parser) parseDeploy(base protocol.IERCTransactionBase, ierc20 *IERC20) (*protocol.DeployCommand, error) {
	maxSupply, err := decimal.NewFromString(strings.TrimSpace(ierc20.Max))
	if err != nil {
		return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "invalid max_supply")
	}

	decimals, err := strconv.ParseInt(strings.TrimSpace(ierc20.Dec), 10, 64)
	if err != nil {
		return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "invalid decimals")
	}

	limit, err := decimal.NewFromString(strings.TrimSpace(ierc20.Lim))
	if err != nil {
		return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "invalid limit")
	}

	wLimit, err := decimal.NewFromString(strings.TrimSpace(ierc20.Wlim))
	if err != nil {
		return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "invalid wallet_limit")
	}

	tick := strings.TrimSpace(ierc20.Tick)
	nonce, err := parser.parseNonce(tick, ierc20.Nonce)
	if err != nil {
		return nil, err
	}

	return &protocol.DeployCommand{
		IERCTransactionBase: base,
		Tick:                strings.TrimSpace(ierc20.Tick),
		MaxSupply:           maxSupply,
		Decimals:            decimals,
		MintLimitOfSingleTx: limit,
		MintLimitOfWallet:   wLimit,
		Workc:               ierc20.Workc,
		Nonce:               nonce.String(),
	}, nil
}

func (parser *IERC20Parser) parseMint(base protocol.IERCTransactionBase, ierc20 *IERC20) (*protocol.MintCommand, error) {

	amount, err := decimal.NewFromString(strings.TrimSpace(ierc20.Amt))
	if err != nil {
		return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "invalid amount")
	}

	tick := strings.TrimSpace(ierc20.Tick)
	nonce, err := parser.parseNonce(tick, ierc20.Nonce)
	if err != nil {
		return nil, err
	}

	if nonce.Equal(decimal.Zero) {
		return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "invalid nonce")
	}

	return &protocol.MintCommand{
		IERCTransactionBase: base,
		Tick:                tick,
		Amount:              amount,
		Nonce:               nonce.String(),
	}, nil
}

func (parser *IERC20Parser) parseTransfer(base protocol.IERCTransactionBase, ierc20 *IERC20) (*protocol.TransferCommand, error) {
	var tick = strings.TrimSpace(ierc20.Tick)

	var records = make([]*protocol.TransferRecord, 0, len(ierc20.To))
	for _, to := range ierc20.To {

		amount, err := decimal.NewFromString(strings.TrimSpace(fmt.Sprintf("%v", to.Amt)))
		if err != nil {
			return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "invalid amount")
		}

		records = append(records, &protocol.TransferRecord{
			Protocol: base.Protocol,
			Operate:  base.Operate,
			Tick:     tick,
			From:     base.From,
			Recv:     strings.ToLower(strings.TrimSpace(to.Recv)),
			Amount:   amount,
		})
	}

	return &protocol.TransferCommand{IERCTransactionBase: base, Records: records}, nil
}

func (parser *IERC20Parser) parseFreezeSell(base protocol.IERCTransactionBase, ierc20 *IERC20) (*protocol.FreezeSellCommand, error) {
	var tick = strings.TrimSpace(ierc20.Tick)

	var records = make([]protocol.FreezeRecord, 0, len(ierc20.Freeze))
	for _, freeze := range ierc20.Freeze {

		tick = strings.TrimSpace(freeze.Tick)
		nonce, err := parser.parseNonce(tick, freeze.Nonce)
		if err != nil {
			return nil, err
		}

		amount, err := decimal.NewFromString(strings.TrimSpace(fmt.Sprintf("%v", freeze.Amt)))
		if err != nil {
			return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "invalid amount")
		}

		ethValue, err := decimal.NewFromString(strings.TrimSpace(freeze.Value))
		if err != nil {
			return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, fmt.Sprintf("invalid sell value. %s", freeze.Value))
		}

		gasPrice, err := decimal.NewFromString(freeze.GasPrice)
		if err != nil {
			return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "invalid gas_price")
		}

		records = append(records, protocol.FreezeRecord{
			Protocol:   base.Protocol,
			Operate:    base.Operate,
			Tick:       tick,
			Platform:   strings.ToLower(strings.TrimSpace(freeze.Platform)),
			Seller:     strings.ToLower(strings.TrimSpace(freeze.Seller)),
			SellerSign: strings.TrimSpace(freeze.Sign),
			SignNonce:  nonce.String(),
			Amount:     amount,
			Value:      ethValue,
			GasPrice:   gasPrice,
		})
	}

	return &protocol.FreezeSellCommand{IERCTransactionBase: base, Records: records}, nil
}

func (parser *IERC20Parser) parseUnfreezeSell(base protocol.IERCTransactionBase, ierc20 *IERC20) (*protocol.UnfreezeSellCommand, error) {
	var records = make([]protocol.UnfreezeRecord, 0, len(ierc20.Unfreeze))
	for _, unfreeze := range ierc20.Unfreeze {
		records = append(records, protocol.UnfreezeRecord{
			Protocol:            base.Protocol,
			Operate:             base.Operate,
			TxHash:              strings.ToLower(unfreeze.TxHash),
			PositionInIERC20Txs: unfreeze.PositionInIERC20Txs,
			Sign:                unfreeze.Sign,
			Msg:                 unfreeze.Msg,
		})
	}

	return &protocol.UnfreezeSellCommand{IERCTransactionBase: base, Records: records}, nil
}

func (parser *IERC20Parser) parseProxyTransfer(base protocol.IERCTransactionBase, ierc20 *IERC20) (*protocol.ProxyTransferCommand, error) {
	var records = make([]protocol.ProxyTransferRecord, 0, len(ierc20.Proxy))
	for _, proxy := range ierc20.Proxy {

		tick := strings.TrimSpace(proxy.Tick)

		nonce, err := parser.parseNonce(tick, proxy.Nonce)
		if err != nil {
			return nil, err
		}

		amount, err := decimal.NewFromString(strings.TrimSpace(fmt.Sprintf("%v", proxy.Amt)))
		if err != nil {
			return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "invalid amount")
		}

		ethValue, err := decimal.NewFromString(strings.TrimSpace(proxy.Value))
		if err != nil {
			return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "invalid eth value")
		}

		records = append(records, protocol.ProxyTransferRecord{
			Protocol:    base.Protocol,
			Operate:     base.Operate,
			Tick:        tick,
			From:        strings.ToLower(strings.TrimSpace(proxy.From)),
			To:          strings.ToLower(strings.TrimSpace(proxy.To)),
			Amount:      amount,
			Value:       ethValue,
			Sign:        strings.TrimSpace(proxy.Sign),
			SignerNonce: nonce.String(),
		})
	}

	return &protocol.ProxyTransferCommand{IERCTransactionBase: base, Records: records}, nil
}

func (parser *IERC20Parser) parseConfigStaking(base protocol.IERCTransactionBase, data []byte) (*protocol.ConfigStakeCommand, error) {
	var e ConfigStaking
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, protocol.NewProtocolError(protocol.InvalidProtocolFormat, "invalid protocol format")
	}

	var details []*protocol.TickConfigDetail
	for _, item := range e.Details {
		details = append(details, &protocol.TickConfigDetail{
			Tick:                 item.Tick,
			RewardsRatioPerBlock: item.RewardsRatioPerBlock,
			MaxAmount:            item.MaxAmount,
		})
	}

	return &protocol.ConfigStakeCommand{
		IERCTransactionBase: base,
		Pool:                strings.ToLower(e.Pool),
		PoolSubID:           uint64(e.PoolSubID),
		Owner:               base.From,
		Admins:              []string{strings.ToLower(e.Owner)},
		Name:                e.Name,
		StopBlock:           uint64(e.StopBlock),
		Details:             details,
	}, nil
}

func (parser *IERC20Parser) parseStakingOrUnStaking(base protocol.IERCTransactionBase, data []byte) (*protocol.StakingCommand, error) {
	var e Staking
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, protocol.NewProtocolError(protocol.InvalidProtocolFormat, "invalid protocol format")
	}

	var (
		pool      = strings.ToLower(e.Pool)
		poolSubID = uint64(e.PoolSubID)
	)

	var records []*protocol.StakingDetail
	for _, item := range e.Details {
		records = append(records, &protocol.StakingDetail{
			Protocol:  base.Protocol,
			Operate:   base.Operate,
			Staker:    base.From,
			Pool:      pool,
			PoolSubID: poolSubID,
			Tick:      item.Tick,
			Amount:    item.Amount,
		})
	}

	return &protocol.StakingCommand{
		IERCTransactionBase: base,
		Pool:                pool,
		PoolSubID:           poolSubID,
		Details:             records,
	}, nil
}

func (parser *IERC20Parser) parseProxyUnStaking(base protocol.IERCTransactionBase, data []byte) (*protocol.StakingCommand, error) {
	var e Staking
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, protocol.NewProtocolError(protocol.InvalidProtocolFormat, "invalid protocol format")
	}

	var (
		pool      = strings.ToLower(e.Pool)
		poolSubID = uint64(e.PoolSubID)
	)

	var records []*protocol.StakingDetail
	for _, item := range e.Details {
		records = append(records, &protocol.StakingDetail{
			Protocol:  base.Protocol,
			Operate:   base.Operate,
			Staker:    strings.ToLower(item.Staker),
			Pool:      pool,
			PoolSubID: poolSubID,
			Tick:      item.Tick,
			Amount:    item.Amount,
		})
	}

	return &protocol.StakingCommand{
		IERCTransactionBase: base,
		Pool:                pool,
		PoolSubID:           poolSubID,
		Details:             records,
	}, nil

}
