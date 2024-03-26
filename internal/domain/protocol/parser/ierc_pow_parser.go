package parser

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"github.com/IErcOrg/IERC_Indexer/internal/domain"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
	"github.com/shopspring/decimal"
)

// IERC-PoW
type (
	// ========= deploy & mint =========

	Rule struct {
		Pow               decimal.Decimal `json:"pow"`
		MinWorkc          string          `json:"min_workc"`
		DifficultyRatio   decimal.Decimal `json:"difficulty_ratio"`
		Pos               decimal.Decimal `json:"pos"`
		Pool              string          `json:"pool"`
		MaxRewardBlockNum string          `json:"max_reward_block,omitempty"`
	}

	IERCPoWDeploy struct {
		Tick       string                     `json:"tick"`
		Max        decimal.Decimal            `json:"max"`
		Dec        string                     `json:"dec"`
		Tokenomics map[string]decimal.Decimal `json:"tokenomics"`
		Rule       Rule                       `json:"rule"`
	}

	IERCPoWMint struct {
		Tick     string `json:"tick"`
		UsePoint string `json:"use_point,omitempty"`
		Block    string `json:"block,omitempty"`
		Nonce    string `json:"nonce"`
	}

	IERCPoWMidify struct {
		Tick      string          `json:"tick"`
		MaxSupply decimal.Decimal `json:"max"`
	}

	IERCPoWAirdropClaim struct {
		Tick        string          `json:"tick"`
		ClaimAmount decimal.Decimal `json:"claim"`
	}

	// ========== transfer ==========

	IERCPoWTransferRecord struct {
		Recv   string          `json:"recv"`
		Amount decimal.Decimal `json:"amt"`
	}
	IERCPoWTransfer struct {
		Tick    string                   `json:"tick"`
		Records []*IERCPoWTransferRecord `json:"to"`
	}

	// ========== freeze & unfreeze & proxy_transfer ==========

	IERCPoWFreezeRecord struct {
		Tick     string          `json:"tick"`
		Platform string          `json:"platform"`
		Seller   string          `json:"seller"`
		Amt      decimal.Decimal `json:"amt"`
		Value    decimal.Decimal `json:"value"`
		GasPrice decimal.Decimal `json:"gasPrice"`
		Sign     string          `json:"sign"`
		Nonce    string          `json:"nonce"`
	}
	IERCPoWFreeze struct {
		Records []*IERCPoWFreezeRecord `json:"freeze"`
	}

	IERCPoWUnfreezeRecord struct {
		TxHash              string `json:"txHash"`
		PositionInIERC20Txs string `json:"position"`
		Sign                string `json:"sign"`
		Msg                 string `json:"msg"`
	}
	IERCPoWUnfreeze struct {
		Records []*IERCPoWUnfreezeRecord `json:"unfreeze"`
	}

	IERCPoWProxyTransferRecord struct {
		Tick   string          `json:"tick"`
		From   string          `json:"from"`
		To     string          `json:"to"`
		Amount decimal.Decimal `json:"amt"`
		Value  decimal.Decimal `json:"value"`
		Sign   string          `json:"sign"`
		Nonce  string          `json:"nonce"`
	}
	IERCPoWProxyTransfer struct {
		Records []*IERCPoWProxyTransferRecord `json:"proxy"`
	}
)

type IERCPoWParser struct {
	headerLength int

	supportedAirDropTicks map[string]struct{} //
}

func newIERC20PoWParser(header string) Parser {
	p := &IERCPoWParser{
		headerLength:          len(header),
		supportedAirDropTicks: make(map[string]struct{}),
	}

	p.supportedAirDropTicks["ethpi"] = struct{}{}

	return p
}

func (parser *IERCPoWParser) CheckFormat(_ []byte) error {
	return nil
}

func (parser *IERCPoWParser) Parse(tx *domain.Transaction) (protocol.IERCTransaction, error) {

	var base protocol.IERCTransactionBase
	data := []byte(tx.TxData[parser.headerLength:])
	if err := json.Unmarshal(data, &base); err != nil {
		return nil, protocol.NewProtocolError(protocol.InvalidProtocolFormat, "invalid protocol format")
	}

	base = protocol.IERCTransactionBase{
		BlockNumber:        tx.BlockNumber,
		TxHash:             tx.Hash,
		TxValue:            tx.TxValue,
		PositionInBlockTxs: tx.PositionInTxs,
		From:               strings.ToLower(tx.From),
		To:                 strings.ToLower(tx.To),
		Gas:                tx.Gas,
		GasPrice:           tx.GasPrice,
		EventAt:            tx.CreatedAt,
		Protocol:           base.Protocol,
		Operate:            base.Operate,
	}

	if err := base.Validate(); err != nil {
		return nil, err
	}

	switch base.Operate {
	case protocol.OpDeploy:
		return parser.parseDeploy(base, data)

	case protocol.OpMint:
		return parser.parseMint(base, data)

	case protocol.OpTransfer:
		return parser.parseTransfer(base, data)

	case protocol.OpFreezeSell:
		return parser.parseFreezeSell(base, data)

	case protocol.OpUnfreezeSell:
		return parser.parseUnfreezeSell(base, data)

	case protocol.OpProxyTransfer:
		return parser.parseProxyTransfer(base, data)

	case protocol.OpPoWModify:
		return parser.parseModify(base, data)

	case protocol.OpPoWClaimAirdrop:
		return parser.parseAirdropClaim(base, data)

	default:
		return nil, protocol.NewProtocolError(protocol.UnknownProtocolOperate, "unknown operate")
	}
}

func (parser *IERCPoWParser) parseDeploy(base protocol.IERCTransactionBase, data []byte) (*protocol.DeployPoWCommand, error) {
	var e IERCPoWDeploy
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}

	var decimals, err = strconv.ParseInt(e.Dec, 10, 64)
	if err != nil {
		return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "invalid dec")
	}

	var tokenomics []protocol.TokenomicsDetail
	for block, amount := range e.Tokenomics {
		blockNumber, err := strconv.ParseUint(block, 10, 64)
		if err != nil {
			return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "invalid tokenomics")
		}
		tokenomics = append(tokenomics, protocol.TokenomicsDetail{BlockNumber: blockNumber, Amount: amount})
	}

	var maxBlockNum uint64
	if e.Rule.MaxRewardBlockNum != "" {
		var err error
		maxBlockNum, err = strconv.ParseUint(e.Rule.MaxRewardBlockNum, 10, 64)
		if err != nil {
			return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "invalid max_reward_block")
		}
	}

	sort.Slice(tokenomics, func(i, j int) bool {
		return tokenomics[i].BlockNumber < tokenomics[j].BlockNumber
	})

	return &protocol.DeployPoWCommand{
		IERCTransactionBase: base,
		Tick:                strings.TrimSpace(e.Tick),
		Decimals:            decimals,
		MaxSupply:           e.Max,
		TokenomicsDetails:   tokenomics,
		DistributionRule: protocol.DistributionRule{
			PowRatio:          e.Rule.Pow,
			MinWorkC:          e.Rule.MinWorkc,
			DifficultyRatio:   e.Rule.DifficultyRatio,
			PosRatio:          e.Rule.Pos,
			PosPool:           e.Rule.Pool,
			MaxRewardBlockNum: maxBlockNum,
		},
	}, nil
}

func (parser *IERCPoWParser) parseMint(base protocol.IERCTransactionBase, data []byte) (*protocol.MintPoWCommand, error) {
	var e IERCPoWMint
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}

	var point decimal.Decimal
	if e.UsePoint != "" {
		var err error
		point, err = decimal.NewFromString(e.UsePoint)
		if err != nil {
			return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "invalid point")
		}
	}

	var block uint64
	if e.Block != "" {
		var err error
		block, err = strconv.ParseUint(e.Block, 10, 64)
		if err != nil {
			return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "invalid block")
		}

		if block >= protocol.DPoSDisableDualMiningBlockHeight {
			point = decimal.Zero
		}
	}

	nonce, err := strconv.ParseUint(e.Nonce, 10, 64)
	if err != nil {
		return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "invalid nonce")
	}

	return protocol.NewMintPoWCommand(base, e.Tick, point, block, nonce), nil
}

func (parser *IERCPoWParser) parseTransfer(base protocol.IERCTransactionBase, data []byte) (*protocol.TransferCommand, error) {
	var e IERCPoWTransfer
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}

	var records = make([]*protocol.TransferRecord, 0, len(e.Records))
	for _, record := range e.Records {
		records = append(records, &protocol.TransferRecord{
			Protocol: base.Protocol,
			Operate:  base.Operate,
			Tick:     e.Tick,
			From:     base.From,
			Recv:     record.Recv,
			Amount:   record.Amount,
		})
	}

	return &protocol.TransferCommand{IERCTransactionBase: base, Records: records}, nil
}

func (parser *IERCPoWParser) parseFreezeSell(base protocol.IERCTransactionBase, data []byte) (*protocol.FreezeSellCommand, error) {
	var e IERCPoWFreeze
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}

	var records = make([]protocol.FreezeRecord, 0, len(e.Records))
	for _, record := range e.Records {
		records = append(records, protocol.FreezeRecord{
			Protocol:   base.Protocol,
			Operate:    base.Operate,
			Tick:       record.Tick,
			Platform:   strings.ToLower(record.Platform),
			Seller:     strings.ToLower(record.Seller),
			SellerSign: record.Sign,
			SignNonce:  record.Nonce,
			Amount:     record.Amt,
			Value:      record.Value,
			GasPrice:   record.GasPrice,
		})
	}

	return &protocol.FreezeSellCommand{IERCTransactionBase: base, Records: records}, nil
}

func (parser *IERCPoWParser) parseUnfreezeSell(base protocol.IERCTransactionBase, data []byte) (*protocol.UnfreezeSellCommand, error) {
	var e IERCPoWUnfreeze
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}

	var records = make([]protocol.UnfreezeRecord, 0, len(e.Records))
	for _, record := range e.Records {
		positionInIERC20Txs, err := strconv.ParseInt(record.PositionInIERC20Txs, 10, 64)
		if err != nil {
			return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "invalid position")
		}

		records = append(records, protocol.UnfreezeRecord{
			Protocol:            base.Protocol,
			Operate:             base.Operate,
			TxHash:              strings.ToLower(record.TxHash),
			PositionInIERC20Txs: int32(positionInIERC20Txs),
			Sign:                record.Sign,
			Msg:                 record.Msg,
		})
	}

	return &protocol.UnfreezeSellCommand{IERCTransactionBase: base, Records: records}, nil
}

func (parser *IERCPoWParser) parseProxyTransfer(base protocol.IERCTransactionBase, data []byte) (*protocol.ProxyTransferCommand, error) {
	var e IERCPoWProxyTransfer
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}

	var records = make([]protocol.ProxyTransferRecord, 0, len(e.Records))
	for _, record := range e.Records {
		records = append(records, protocol.ProxyTransferRecord{
			Protocol:    base.Protocol,
			Operate:     base.Operate,
			Tick:        record.Tick,
			From:        strings.ToLower(record.From),
			To:          strings.ToLower(record.To),
			Amount:      record.Amount,
			Value:       record.Value,
			Sign:        record.Sign,
			SignerNonce: record.Nonce,
		})
	}

	return &protocol.ProxyTransferCommand{IERCTransactionBase: base, Records: records}, nil
}

func (parser *IERCPoWParser) supportAirdrop(tick string) bool {
	_, ok := parser.supportedAirDropTicks[tick]
	return ok
}

func (parser *IERCPoWParser) parseModify(base protocol.IERCTransactionBase, data []byte) (*protocol.ModifyCommand, error) {
	var e IERCPoWMidify
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}

	if !parser.supportAirdrop(e.Tick) {
		return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "not support airdrop")
	}

	return &protocol.ModifyCommand{
		IERCTransactionBase: base,
		Tick:                e.Tick,
		MaxSupply:           e.MaxSupply,
	}, nil
}

func (parser *IERCPoWParser) parseAirdropClaim(base protocol.IERCTransactionBase, data []byte) (*protocol.ClaimAirdropCommand, error) {

	var e IERCPoWAirdropClaim
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}

	if !parser.supportAirdrop(e.Tick) {
		return nil, protocol.NewProtocolError(protocol.InvalidProtocolParams, "not support airdrop")
	}

	return &protocol.ClaimAirdropCommand{
		IERCTransactionBase: base,
		Tick:                e.Tick,
		ClaimAmount:         e.ClaimAmount,
	}, nil
}
