package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/IErcOrg/IERC_Indexer/internal/domain/balance"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/staking"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/tick"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

type AggregateRoot struct {
	// initialize state
	PreviousBlock uint64
	Block         *Block
	TicksMap      map[string]tick.Tick
	BalancesMap   map[balance.BalanceKey]*balance.Balance
	Signatures    map[string]*IERC20TransferredEvent
	StakingPools  map[string]*staking.PoolAggregate

	// config
	invalidTxHashMap map[string]struct{}
	feeStartBlock    uint64

	// runtime
	mintFlag map[string]struct{}
	Events   []Event
}

func NewBlockAggregate(previous uint64, block *Block, invalidTxHashMap map[string]struct{}, feeStartBlock uint64) *AggregateRoot {
	if invalidTxHashMap == nil {
		invalidTxHashMap = make(map[string]struct{})
	}

	return &AggregateRoot{
		PreviousBlock:    previous,
		Block:            block,
		TicksMap:         make(map[string]tick.Tick),
		BalancesMap:      make(map[balance.BalanceKey]*balance.Balance),
		Signatures:       make(map[string]*IERC20TransferredEvent),
		StakingPools:     make(map[string]*staking.PoolAggregate),
		invalidTxHashMap: invalidTxHashMap,
		feeStartBlock:    feeStartBlock,
		mintFlag:         make(map[string]struct{}),
		Events:           nil,
	}
}

func (root *AggregateRoot) checkTxHash(txHash string) error {
	if _, existed := root.invalidTxHashMap[txHash]; existed {
		return protocol.NewProtocolError(protocol.InvalidTxHash, "invalid tx hash")
	}

	return nil
}

func (root *AggregateRoot) getOrCreateBalance(address, tick string) *balance.Balance {
	key := balance.NewBalanceKey(address, tick)
	entity, existed := root.BalancesMap[key]
	if existed {
		return entity
	}

	entity = &balance.Balance{
		ID:               0,
		Address:          address,
		Tick:             tick,
		Available:        decimal.Zero,
		Freeze:           decimal.Zero,
		MintedAmount:     decimal.Zero,
		LastUpdatedBlock: 0,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	root.BalancesMap[key] = entity

	return entity
}

func (root *AggregateRoot) isMinted(address, tick string) bool {
	key := fmt.Sprintf("%s-%s", address, tick)
	_, existed := root.mintFlag[key]
	return existed
}

func (root *AggregateRoot) markMinted(address, tick string) {
	key := fmt.Sprintf("%s-%s", address, tick)
	root.mintFlag[key] = struct{}{}
}

func (root *AggregateRoot) resetMintFlag() {
	root.mintFlag = make(map[string]struct{})
}

type totalShare struct {
	PoWTotalShare decimal.Decimal
	PoSTotalShare decimal.Decimal
}

func (t *totalShare) AddPoSShare(share decimal.Decimal) {
	t.PoSTotalShare = t.PoSTotalShare.Add(share)
}

func (t *totalShare) AddPoWShare(share decimal.Decimal) {
	t.PoWTotalShare = t.PoWTotalShare.Add(share)
}

func (root *AggregateRoot) calculatePoWMintShare() map[string]*totalShare {

	shares := make(map[string]*totalShare)

	for _, transaction := range root.Block.Transactions {

		if transaction.IsProcessed || transaction.IERCTransaction == nil {
			continue
		}

		command, ok := transaction.IERCTransaction.(*protocol.MintPoWCommand)
		if !ok {
			continue
		}

		tickName := command.Tick()

		t, existed := root.TicksMap[tickName]
		if !existed {
			continue
		}

		tickEntity, ok := t.(*tick.IERCPoWTick)
		if !ok || tickEntity.Protocol != command.Protocol {
			continue
		}

		pool, err := root.getPoolAggregate(tickEntity.Rule.PosPool)
		if err != nil {
			panic(err)
		}

		ts, existed := shares[tickName]
		if !existed {
			ts = new(totalShare)
			shares[tickName] = ts
		}

		if root.isMinted(command.From, tickName) {

			transaction.Code = int32(protocol.MintErrTickMinted)
			transaction.Remark = "has been minted"
			transaction.IsProcessed = true
			transaction.UpdatedAt = time.Now()
			continue
		}

		var canMint bool

		switch {
		case command.IsDPoS() && command.IsPoW():

			share := root.calcPoWShare(command, tickEntity)
			if share.IsZero() {
				continue
			}

			points := command.Points()

			if command.BlockNumber > protocol.DPoSMintMintPointsLimitBlockHeight &&
				points.LessThan(decimal.NewFromInt(protocol.DPoSMintMinPoints)) {
				points = decimal.Zero
			}

			if !pool.CanUseRewards(command.BlockNumber, command.From, points) {
				continue
			}

			ts.AddPoWShare(share)
			ts.AddPoSShare(points)
			canMint = true

		case command.IsDPoS():

			points := command.Points()
			if command.BlockNumber > protocol.DPoSMintMintPointsLimitBlockHeight &&
				points.LessThan(decimal.NewFromInt(protocol.DPoSMintMinPoints)) {
				continue
			}

			if !pool.CanUseRewards(command.BlockNumber, command.From, points) {
				continue
			}

			ts.AddPoSShare(points)
			canMint = true

		case command.IsPoW():
			share := root.calcPoWShare(command, tickEntity)
			if share.IsZero() {
				continue
			}

			ts.AddPoWShare(share)
			canMint = true
		}

		if canMint {
			root.markMinted(command.From, tickName)
		}
	}

	root.resetMintFlag()

	return shares
}

func (root *AggregateRoot) calcPoWShare(tx *protocol.MintPoWCommand, tickEntity *tick.IERCPoWTick) decimal.Decimal {

	var diff = max(tx.Block(), tx.BlockNumber) - min(tx.Block(), tx.BlockNumber)
	if diff > 5 {
		return decimal.Zero
	}

	return tickEntity.CalculateMintShareBasedOnHash(tx.BlockNumber, tx.TxHash)
}

func (root *AggregateRoot) Handle() {

	shares := root.calculatePoWMintShare()

	for _, transaction := range root.Block.Transactions {
		if transaction.IsProcessed {
			continue
		}

		transaction.IsProcessed = true
		transaction.UpdatedAt = time.Now()

		if transaction.IERCTransaction == nil {
			continue
		}

		var err error
		switch tx := transaction.IERCTransaction.(type) {
		case *protocol.DeployCommand:
			err = root.HandleDeploy(tx)

		case *protocol.MintCommand:
			err = root.HandleMint(tx)

		case *protocol.DeployPoWCommand:
			err = root.handleDeployPow(tx)

		case *protocol.MintPoWCommand:
			powMintTotalShare, posMintTotalShare := decimal.Zero, decimal.Zero
			if ts, existed := shares[tx.Tick()]; existed {
				powMintTotalShare = ts.PoWTotalShare
				posMintTotalShare = ts.PoSTotalShare
			}
			err = root.handleMintPoW(tx, powMintTotalShare, posMintTotalShare)

		case *protocol.ModifyCommand:
			err = root.handleModify(tx)

		case *protocol.ClaimAirdropCommand:
			err = root.handleClaimAirdrop(tx)

		case *protocol.TransferCommand:
			err = root.HandleTransfer(tx)

		case *protocol.UnfreezeSellCommand:
			err = root.HandleUnfreezeSell(tx)

		case *protocol.FreezeSellCommand:
			err = root.HandleFreezeSell(tx)

		case *protocol.ProxyTransferCommand:
			err = root.HandleProxyTransfer(tx)

		case *protocol.ConfigStakeCommand:
			err = root.handleConfigStaking(tx)

		case *protocol.StakingCommand:
			switch tx.Operate {
			case protocol.OpStaking:
				err = root.handleStaking(tx)
			case protocol.OpUnStaking:
				err = root.handleUnStaking(tx)
			case protocol.OpProxyUnStaking:
				err = root.handleProxyUnStaking(tx)
			}
		}

		if err != nil {
			var pErr *protocol.ProtocolError
			if errors.As(err, &pErr) {
				transaction.Code = pErr.Code()
				transaction.Remark = pErr.Message()
			} else {
				transaction.Code = int32(protocol.UnknownError)
				transaction.Remark = err.Error()
			}
			continue
		}
	}
}

// ==================== about tick: deploy & mint ====================

func (root *AggregateRoot) HandleDeploy(command *protocol.DeployCommand) (err error) {

	event := &IERC20TickCreatedEvent{
		BlockNumber:       command.BlockNumber,
		PrevBlockNumber:   root.PreviousBlock,
		TxHash:            command.TxHash,
		PositionInIERCTxs: 0,
		From:              command.From,
		To:                command.To,
		Value:             command.TxValue.String(),
		Data: &IERC20TickCreated{
			Protocol:    command.Protocol,
			Operate:     command.Operate,
			Tick:        command.Tick,
			Decimals:    command.Decimals,
			MaxSupply:   command.MaxSupply,
			Limit:       command.MintLimitOfSingleTx,
			WalletLimit: command.MintLimitOfWallet,
			WorkC:       command.Workc,
			Nonce:       command.Nonce,
		},
		ErrCode:   0,
		ErrReason: "",
		EventAt:   command.EventAt,
	}
	defer func() {
		event.SetError(err)
		root.Events = append(root.Events, event)
	}()

	if _, existed := root.TicksMap[command.Tick]; existed {
		return protocol.NewProtocolError(protocol.TickExited, "tick already existed")
	}

	root.TicksMap[command.Tick] = tick.NewTickFromDeployCommand(command)

	return nil
}

func (root *AggregateRoot) HandleMint(command *protocol.MintCommand) (err error) {

	if err = root.checkTxHash(command.TxHash); err != nil {
		return
	}

	ee := &IERC20MintedEvent{
		BlockNumber:       command.BlockNumber,
		PrevBlockNumber:   root.PreviousBlock,
		TxHash:            command.TxHash,
		PositionInIERCTxs: 0,
		From:              command.From,
		To:                command.To,
		Value:             command.TxValue.String(),
		Data: &IERC20Minted{
			Protocol:     command.Protocol,
			Operate:      command.Operate,
			Tick:         command.Tick,
			From:         protocol.ZeroAddress,
			To:           command.From,
			MintedAmount: command.Amount,
			Gas:          command.Gas,
			GasPrice:     command.GasPrice,
			Nonce:        command.Nonce,
		},
		ErrCode:   0,
		ErrReason: "",
		EventAt:   command.EventAt,
	}

	defer func() {
		ee.SetError(err)
		root.Events = append(root.Events, ee)
	}()

	tickEntity, existed := root.TicksMap[command.Tick]
	if !existed {
		return protocol.NewProtocolError(protocol.TickNotExist, "tick not existed")
	}

	if root.isMinted(command.From, command.Tick) {
		return protocol.NewProtocolError(protocol.MintErrTickMinted, "has been minted")
	}

	ierc20TickEntity, ok := tickEntity.(*tick.IERC20Tick)
	if !ok {
		return protocol.NewProtocolError(protocol.MintErrTickProtocolNoMatch, "tick protocol no match")
	}

	if err = ierc20TickEntity.ValidateHash(command.TxHash); err != nil {
		return err
	}

	minerBalance := root.getOrCreateBalance(command.From, command.Tick)
	if err = ierc20TickEntity.CanMint(command.Amount, minerBalance.MintedAmount); err != nil {
		return err
	}

	ierc20TickEntity.Mint(command.BlockNumber, command.Amount)
	minerBalance.AddMint(command.BlockNumber, command.Amount)
	root.markMinted(command.From, command.Tick)

	return
}

func (root *AggregateRoot) handleDeployPow(command *protocol.DeployPoWCommand) (err error) {

	event := &IERCPoWTickCreatedEvent{
		BlockNumber:       command.BlockNumber,
		PrevBlockNumber:   root.PreviousBlock,
		TxHash:            command.TxHash,
		PositionInIERCTxs: 0,
		From:              command.From,
		To:                command.To,
		Value:             command.TxValue.String(),
		Data: &IERCPoWTickCreated{
			Protocol:   command.Protocol,
			Operate:    command.Operate,
			Tick:       command.Tick,
			Decimals:   command.Decimals,
			MaxSupply:  command.MaxSupply,
			Tokenomics: command.TokenomicsDetails,
			Rule:       command.DistributionRule,
			Creator:    command.From,
		},
		ErrCode:   0,
		ErrReason: "",
		EventAt:   command.EventAt,
	}
	defer func() {
		event.SetError(err)
		root.Events = append(root.Events, event)
	}()

	if _, existed := root.TicksMap[command.Tick]; existed {
		return protocol.NewProtocolError(protocol.TickExited, "tick already existed")
	}

	if _, err = root.getPoolAggregate(command.DistributionRule.PosPool); err != nil {
		return err
	}

	root.TicksMap[command.Tick] = tick.NewIERCPoWTickFromDeployCommand(command)

	return
}

func (root *AggregateRoot) handleMintPoW(command *protocol.MintPoWCommand, powTotalShare decimal.Decimal, posTotalShare decimal.Decimal) (err error) {

	ee := &IERCPoWMintedEvent{
		BlockNumber:       command.BlockNumber,
		PrevBlockNumber:   root.PreviousBlock,
		TxHash:            command.TxHash,
		PositionInIERCTxs: 0,
		From:              command.From,
		To:                command.To,
		Value:             command.TxValue.String(),
		Data: &IERCPoWMinted{
			Protocol:        command.Protocol,
			Operate:         command.Operate,
			Tick:            command.Tick(),
			From:            protocol.ZeroAddress,
			To:              command.From,
			IsPoW:           command.IsPoW(),
			PoWTotalShare:   powTotalShare,
			PoWMinerShare:   decimal.Zero,
			PoWMintedAmount: decimal.Zero,
			IsPoS:           command.IsDPoS(),
			PoSTotalShare:   posTotalShare,
			PoSMinerShare:   decimal.Zero,
			PoSMintedAmount: decimal.Zero,
			PoSPointsSource: "",
			Gas:             command.Gas,
			GasPrice:        command.GasPrice,
			Nonce:           command.Nonce(),
		},
		ErrCode:   0,
		ErrReason: "",
		EventAt:   command.EventAt,
	}
	root.Events = append(root.Events, ee)

	defer func() {
		ee.SetError(err)
	}()

	tickName := command.Tick()

	if root.isMinted(command.From, tickName) {
		return protocol.NewProtocolError(protocol.MintErrTickMinted, "has been minted")
	}

	t, existed := root.TicksMap[tickName]
	if !existed {
		return protocol.NewProtocolError(protocol.MintErrTickNotFound, "tick not found")
	}

	tickEntity, ok := t.(*tick.IERCPoWTick)
	if !ok || tickEntity.Protocol != command.Protocol {
		return protocol.NewProtocolError(protocol.MintErrTickNotSupportPoW, "not support pow")
	}

	pool, err := root.getPoolAggregate(tickEntity.Rule.PosPool)
	if err != nil {
		return err
	}

	params := &tick.PoWMintParams{
		CurrentBlock:   command.BlockNumber,
		EffectiveBlock: command.Block(),
		IsPoW:          command.IsPoW(),
		IsDPoS:         command.IsDPoS(),
		TotalPoWShare:  powTotalShare,
		MinerPoWShare:  decimal.Zero,
		TotalPoSShare:  posTotalShare,
		MinerPoSShare:  decimal.Zero,
	}

	switch {
	case command.IsDPoS() && command.IsPoW():

		params.MinerPoWShare = tickEntity.CalculateMintShareBasedOnHash(command.BlockNumber, command.TxHash)
		if params.MinerPoWShare.IsZero() {
			return protocol.NewProtocolError(protocol.MintErrPoWShareZero, "invalid pow mint")
		}

		points := command.Points()

		if command.BlockNumber > protocol.DPoSMintMintPointsLimitBlockHeight &&
			points.LessThan(decimal.NewFromInt(protocol.DPoSMintMinPoints)) {
			points = decimal.Zero
			params.IsDPoS = false
		}

		if !pool.CanUseRewards(command.BlockNumber, command.From, points) {
			return protocol.NewProtocolError(protocol.UseRewardsErrRewardsInsufficient, "point insufficient")
		}

		params.MinerPoSShare = points

	case command.IsDPoS():
		points := command.Points()

		if command.BlockNumber > protocol.DPoSMintMintPointsLimitBlockHeight && points.LessThan(decimal.NewFromInt(protocol.DPoSMintMinPoints)) {
			return protocol.NewProtocolError(protocol.MintErrDPoSMintPointsTooLow, "point too low")
		}

		if !pool.CanUseRewards(command.BlockNumber, command.From, points) {
			return protocol.NewProtocolError(protocol.UseRewardsErrRewardsInsufficient, "point insufficient")
		}

		params.MinerPoSShare = points

	case command.IsPoW():
		params.MinerPoWShare = tickEntity.CalculateMintShareBasedOnHash(command.BlockNumber, command.TxHash)
	}

	if err = tickEntity.CanMint(params); err != nil {
		return err
	}

	if err := params.Validate(); err != nil {
		log.Errorf("mint params error. command: %v, params: %v", command, params)
		panic(err)
	}

	if command.IsDPoS() && !params.MinerPoSShare.IsZero() {

		if err = pool.UseRewards(command.BlockNumber, command.From, params.MinerPoSShare); err != nil {
			return err
		}
	}

	powMintedAmount, posMintedAmount := tickEntity.Mint(params)

	minerBalance := root.getOrCreateBalance(command.From, tickName)
	minerBalance.AddMint(command.BlockNumber, powMintedAmount.Add(posMintedAmount))

	ee.Data.PoWMinerShare = params.MinerPoWShare
	ee.Data.PoWMintedAmount = powMintedAmount
	ee.Data.PoSMinerShare = command.Points()
	ee.Data.PoSPointsSource = pool.PoolAddress
	ee.Data.PoSMintedAmount = posMintedAmount

	powBurnAmount, posBurnAmount := tickEntity.Burn()
	burnAmount := powBurnAmount.Add(posBurnAmount)

	if burnAmount.GreaterThan(decimal.Zero) {
		burnEvent := &IERC20TransferredEvent{
			BlockNumber:       command.BlockNumber,
			PrevBlockNumber:   root.PreviousBlock,
			TxHash:            command.TxHash,
			PositionInIERCTxs: 1,
			From:              command.From,
			To:                command.To,
			Value:             command.TxValue.String(),
			Data: &IERC20Transferred{
				Protocol: command.Protocol,
				Operate:  command.Operate,
				Tick:     tickEntity.Tick,
				From:     protocol.ZeroAddress,
				To:       protocol.ZeroAddress,
				Amount:   burnAmount,
			},
			ErrCode:   0,
			ErrReason: "",
			EventAt:   command.EventAt,
		}

		blackHoleBalance := root.getOrCreateBalance(protocol.ZeroAddress, tickEntity.Tick)
		blackHoleBalance.AddAvailable(root.Block.Number, burnAmount)

		root.Events = append(root.Events, burnEvent)
	}

	root.markMinted(command.From, tickName)

	return
}

func (root *AggregateRoot) handleModify(command *protocol.ModifyCommand) (err error) {

	event := &IERCPoWTickCreatedEvent{
		BlockNumber:       command.BlockNumber,
		PrevBlockNumber:   root.PreviousBlock,
		TxHash:            command.TxHash,
		PositionInIERCTxs: 0,
		From:              command.From,
		To:                command.To,
		Value:             command.TxValue.String(),
		Data: &IERCPoWTickCreated{
			Protocol:  command.Protocol,
			Operate:   command.Operate,
			Tick:      command.Tick,
			MaxSupply: command.MaxSupply,
			Creator:   command.From,
		},
		EventAt: command.EventAt,
	}
	root.Events = append(root.Events, event)
	defer func() {
		event.SetError(err)
	}()

	tickEntity, existed := root.TicksMap[command.Tick]
	if !existed {
		return protocol.NewProtocolError(protocol.TickNotExist, "tick not existed")
	}

	powTickEntity, ok := tickEntity.(*tick.IERCPoWTick)
	if !ok {
		return protocol.NewProtocolError(protocol.ErrTickProtocolNoMatch, "tick protocol no match")
	}

	err = powTickEntity.UpdateMaxSupply(command.BlockNumber, command.From, command.MaxSupply)
	if err != nil {
		switch {
		case errors.Is(err, tick.ErrNoPermission):
			err = protocol.NewProtocolError(protocol.ErrUpdateMaxSupplyNoPermission, err.Error())

		case errors.Is(err, tick.ErrMaxAmountLessThanSupply):
			err = protocol.NewProtocolError(protocol.ErrUpdateAmountLessThanSupply, err.Error())

		default:
			err = protocol.NewProtocolError(protocol.ErrUpdateFailed, err.Error())
		}

		return err
	}

	return nil
}

func (root *AggregateRoot) handleClaimAirdrop(command *protocol.ClaimAirdropCommand) (err error) {

	ee := &IERCPoWMintedEvent{
		BlockNumber:       command.BlockNumber,
		PrevBlockNumber:   root.PreviousBlock,
		TxHash:            command.TxHash,
		PositionInIERCTxs: 0,
		From:              command.From,
		To:                command.To,
		Value:             command.TxValue.String(),
		Data: &IERCPoWMinted{
			Protocol:      command.Protocol,
			Operate:       command.Operate,
			Tick:          command.Tick,
			From:          protocol.ZeroAddress,
			To:            command.From,
			IsAirdrop:     true,
			AirdropAmount: command.ClaimAmount,
		},
		ErrCode:   0,
		ErrReason: "",
		EventAt:   command.EventAt,
	}
	root.Events = append(root.Events, ee)
	defer func() {
		ee.SetError(err)
	}()

	tickEntity, existed := root.TicksMap[command.Tick]
	if !existed {
		return protocol.NewProtocolError(protocol.TickNotExist, "tick not existed")
	}

	powTickEntity, ok := tickEntity.(*tick.IERCPoWTick)
	if !ok {
		return protocol.NewProtocolError(protocol.MintErrTickProtocolNoMatch, "tick protocol no match")
	}

	err = powTickEntity.ClaimAirdrop(command.BlockNumber, command.From, command.ClaimAmount)
	if err != nil {
		switch {
		case errors.Is(err, tick.ErrNoPermission):
			err = protocol.NewProtocolError(protocol.MintErrNoPermissionToClaimAirdrop, err.Error())

		case errors.Is(err, tick.ErrAirdropAmountExceedsRemainSupply):
			err = protocol.NewProtocolError(protocol.MintErrAirdropAmountExceedsRemainSupply, err.Error())

		case errors.Is(err, tick.ErrInvalidAmount):
			err = protocol.NewProtocolError(protocol.MintErrInvalidAirdropAmount, err.Error())

		default:
			err = protocol.NewProtocolError(protocol.MintErrAirdropClaimFailed, err.Error())
		}

		return err
	}

	fromBalance := root.getOrCreateBalance(command.From, command.Tick)
	fromBalance.AddAvailable(command.BlockNumber, command.ClaimAmount)

	return nil
}

// ==================== about tick: transfer ====================

func (root *AggregateRoot) HandleTransfer(command *protocol.TransferCommand) error {

	for idx, record := range command.Records {

		ee := &IERC20TransferredEvent{
			BlockNumber:       command.BlockNumber,
			PrevBlockNumber:   root.PreviousBlock,
			TxHash:            command.TxHash,
			PositionInIERCTxs: idx,
			From:              command.From,
			To:                command.To,
			Value:             command.TxValue.String(),
			Data: &IERC20Transferred{
				Protocol: record.Protocol,
				Operate:  record.Operate,
				Tick:     record.Tick,
				From:     record.From,
				To:       record.Recv,
				Amount:   record.Amount,
			},
			ErrCode:   0,
			ErrReason: "",
			EventAt:   command.EventAt,
		}

		root.Events = append(root.Events, ee)

		if err := root.checkTxHash(command.TxHash); err != nil {
			ee.SetError(err)
			continue
		}

		_, existed := root.TicksMap[record.Tick]
		if !existed {
			err := protocol.NewProtocolError(protocol.TickNotExist, "tick not exist")
			ee.SetError(err)
			continue
		}

		err := root.handleTransferRecord(record)
		if err != nil {
			ee.SetError(err)
			continue
		}
	}

	return nil
}

func (root *AggregateRoot) handleTransferRecord(record *protocol.TransferRecord) error {

	fromBalance := root.getOrCreateBalance(record.From, record.Tick)

	if fromBalance.Available.LessThan(record.Amount) {
		return protocol.NewProtocolError(
			protocol.InsufficientAvailableFunds,
			fmt.Sprintf("insufficient balance. available(%s) < transfer(%s)", fromBalance.Available, record.Amount),
		)
	}

	toBalance := root.getOrCreateBalance(record.Recv, record.Tick)

	fromBalance.SubAvailable(root.Block.Number, record.Amount)
	toBalance.AddAvailable(root.Block.Number, record.Amount)
	return nil
}

// ==================== about trade: freeze & unfreeze & proxy_transfer ====================

func (root *AggregateRoot) HandleFreezeSell(command *protocol.FreezeSellCommand) error {

	if err := root.checkTxHash(command.TxHash); err != nil {
		return err
	}

	buyerRemainEthValue := command.TxValue.Shift(-18)
	for idx, record := range command.Records {

		ee := &IERC20TransferredEvent{
			BlockNumber:       command.BlockNumber,
			PrevBlockNumber:   root.PreviousBlock,
			TxHash:            command.TxHash,
			PositionInIERCTxs: idx,
			From:              command.From,
			To:                command.To,
			Value:             command.TxValue.String(),
			Data: &IERC20Transferred{
				Protocol:    record.Protocol,
				Operate:     record.Operate,
				Tick:        record.Tick,
				From:        record.Seller,
				To:          record.Seller,
				Amount:      record.Amount,
				EthValue:    record.Value,
				GasPrice:    record.GasPrice,
				Nonce:       "",
				SignerNonce: record.SignNonce,
				Sign:        record.SellerSign,
			},
			ErrCode:   0,
			ErrReason: "",
			EventAt:   command.EventAt,
		}

		err := root.handleFreezeRecord(&record, buyerRemainEthValue)

		if err != nil {
			ee.SetError(err)
		} else {
			value := record.Value
			if root.Block.Number > root.feeStartBlock {
				value = value.Mul(protocol.ServiceFee)
			}
			buyerRemainEthValue = buyerRemainEthValue.Sub(value)

			root.Signatures[record.SellerSign] = ee
		}

		root.Events = append(root.Events, ee)
	}

	return nil
}

func (root *AggregateRoot) handleFreezeRecord(record *protocol.FreezeRecord, buyerRemainEthValue decimal.Decimal) error {

	_, existed := root.TicksMap[record.Tick]
	if !existed {
		return protocol.NewProtocolError(protocol.TickNotExist, "tick not exist")
	}

	if err := record.ValidateParams(); err != nil {
		return err
	}

	if err := record.ValidateSignature(); err != nil {
		return err
	}

	if ee, existed := root.Signatures[record.SellerSign]; existed {
		switch ee.Data.Operate {
		case protocol.OpFreezeSell:
			return protocol.NewProtocolError(protocol.SignatureAlreadyUsed, "signature already used. freeze_sell")
		case protocol.OpProxyTransfer:
			return protocol.NewProtocolError(protocol.SignatureAlreadyUsed, "signature already used. proxy_transfer")

		case protocol.OpUnfreezeSell:

		default:
			panic("FreezeSell Sign Error")
		}
	}

	value := record.Value
	if root.Block.Number > root.feeStartBlock {
		value = value.Mul(protocol.ServiceFee)
	}
	if buyerRemainEthValue.LessThan(value) {
		return protocol.NewProtocolError(
			protocol.InsufficientValue,
			fmt.Sprintf("insufficient value. remainEthValue(%s) < sellerValue(%s)", buyerRemainEthValue, record.Value),
		)
	}

	sellerBalance := root.getOrCreateBalance(record.Seller, record.Tick)
	if sellerBalance.Available.LessThan(record.Amount) {
		return protocol.NewProtocolError(
			protocol.InsufficientAvailableFunds,
			fmt.Sprintf("insufficient balance. avaliable(%v) < wantFreeze(%v)", sellerBalance.Available, record.Amount),
		)
	}

	sellerBalance.FreezeBalance(root.Block.Number, record.Amount)

	return nil
}

func (root *AggregateRoot) HandleUnfreezeSell(command *protocol.UnfreezeSellCommand) error {

	for idx, record := range command.Records {

		event, err := root.handleUnfreezeRecord(command.To, &record)

		ee := &IERC20TransferredEvent{
			BlockNumber:       command.BlockNumber,
			PrevBlockNumber:   root.PreviousBlock,
			TxHash:            command.TxHash,
			PositionInIERCTxs: idx,
			From:              command.From,
			To:                command.To,
			Value:             command.TxValue.String(),
			Data:              event,
			ErrCode:           0,
			ErrReason:         "",
			EventAt:           command.EventAt,
		}
		if err != nil {
			ee.SetError(err)
		} else {
			root.Signatures[record.Sign] = ee
		}

		root.Events = append(root.Events, ee)
	}

	return nil
}

func (root *AggregateRoot) handleUnfreezeRecord(unfreezeAddress string, record *protocol.UnfreezeRecord) (*IERC20Transferred, *protocol.ProtocolError) {

	var event = &IERC20Transferred{
		Operate: protocol.OpUnfreezeSell,
	}

	ee, existed := root.Signatures[record.Sign]
	if !existed {
		return event, protocol.NewProtocolError(protocol.SignatureNotExist, "signature not exist")
	}

	tickEntity, existed := root.TicksMap[ee.Data.Tick]
	if !existed {
		return event, protocol.NewProtocolError(protocol.TickNotExist, "tick not exist")
	}

	event = &IERC20Transferred{
		Protocol:    tickEntity.GetProtocol(),
		Operate:     protocol.OpUnfreezeSell,
		Tick:        tickEntity.GetName(),
		From:        ee.To,
		To:          ee.From,
		Amount:      ee.Data.Amount,
		EthValue:    ee.Data.EthValue,
		GasPrice:    ee.Data.GasPrice,
		Nonce:       ee.Data.Nonce,
		SignerNonce: ee.Data.SignerNonce,
		Sign:        ee.Data.Sign,
	}

	switch ee.Data.Operate {

	case protocol.OpFreezeSell:
		if ee.From != unfreezeAddress {
			return event, protocol.NewProtocolError(protocol.SignatureNotMatch, "signature address not match")
		}
		if strings.ToLower(ee.TxHash) != strings.ToLower(record.TxHash) {
			return event, protocol.NewProtocolError(protocol.SignatureNotMatch, "freeze hash not match")
		}

	case protocol.OpProxyTransfer:
		return event, protocol.NewProtocolError(protocol.SignatureAlreadyUsed, "signature already used. proxy_transfer")

	case protocol.OpUnfreezeSell:
		return event, protocol.NewProtocolError(protocol.SignatureAlreadyUsed, "signature already used. unfreeze_sell")

	default:
		panic("signature status error")
	}

	var unfreezeAmount = ee.Data.Amount

	sellerBalance := root.getOrCreateBalance(ee.Data.From, ee.Data.Tick)
	if sellerBalance.Freeze.LessThan(unfreezeAmount) {
		return event, protocol.NewProtocolError(
			protocol.InsufficientFreezeFunds,
			fmt.Sprintf("insufficient freeze funds. freeze(%v) < unfreeze(%v)", sellerBalance.Freeze, unfreezeAmount),
		)
	}

	sellerBalance.UnfreezeBalance(root.Block.Number, unfreezeAmount)

	return event, nil
}

func (root *AggregateRoot) HandleProxyTransfer(command *protocol.ProxyTransferCommand) error {

	if err := root.checkTxHash(command.TxHash); err != nil {
		return err
	}

	buyerRemainEthValue := command.TxValue.Shift(-18)

	for idx, record := range command.Records {

		event, err := root.handleProxyTransferRecord(&record, buyerRemainEthValue)
		ee := &IERC20TransferredEvent{
			BlockNumber:       command.BlockNumber,
			PrevBlockNumber:   root.PreviousBlock,
			TxHash:            command.TxHash,
			PositionInIERCTxs: idx,
			From:              command.From,
			To:                command.To,
			Value:             command.TxValue.String(),
			Data:              event,
			ErrCode:           0,
			ErrReason:         "",
			EventAt:           command.EventAt,
		}
		if err != nil {
			ee.SetError(err)
		} else {
			buyerRemainEthValue = buyerRemainEthValue.Sub(record.Value)
			root.Signatures[record.Sign] = ee
		}

		root.Events = append(root.Events, ee)
	}

	return nil
}

func (root *AggregateRoot) handleProxyTransferRecord(record *protocol.ProxyTransferRecord, buyerRemainEthValue decimal.Decimal) (*IERC20Transferred, error) {

	var event = &IERC20Transferred{
		Protocol:    record.Protocol,
		Operate:     record.Operate,
		Tick:        record.Tick,
		From:        record.From,
		To:          record.To,
		Amount:      record.Amount,
		EthValue:    record.Value,
		GasPrice:    decimal.Zero,
		Nonce:       "",
		SignerNonce: record.SignerNonce,
		Sign:        record.Sign,
	}

	tickEntity, existed := root.TicksMap[record.Tick]
	if !existed {
		return event, protocol.NewProtocolError(protocol.TickNotExist, "tick not exist")
	}

	event.Protocol = tickEntity.GetProtocol()

	if err := record.ValidateParams(); err != nil {
		return event, err.(*protocol.ProtocolError)
	}

	if err := record.ValidateSignature(); err != nil {
		return event, err.(*protocol.ProtocolError)
	}

	ee, existed := root.Signatures[record.Sign]
	if !existed {
		return event, protocol.NewProtocolError(protocol.SignatureNotExist, "freeze sell not exist")
	}

	switch ee.Data.Operate {
	case protocol.OpProxyTransfer:
		return event, protocol.NewProtocolError(protocol.SignatureAlreadyUsed, "signature already used")
	case protocol.OpUnfreezeSell:
		return event, protocol.NewProtocolError(protocol.SignatureAlreadyUsed, "signature already unfreeze")

	case protocol.OpFreezeSell:

	default:
		panic("proxy transfer signature error")
	}

	if root.Block.Number > root.feeStartBlock && buyerRemainEthValue.LessThan(record.Value) {
		return event, protocol.NewProtocolError(
			protocol.InsufficientValue,
			fmt.Sprintf("insufficient value. remainETHValue(%s) < recordValue(%s)", buyerRemainEthValue, record.Value),
		)
	}

	fromBalance := root.getOrCreateBalance(record.From, record.Tick)
	if fromBalance.Freeze.LessThan(record.Amount) {
		return event, protocol.NewProtocolError(
			protocol.InsufficientFreezeFunds,
			fmt.Sprintf("from insufficient balance. freeze(%s) < transfer(%s)", fromBalance.Freeze, record.Amount),
		)
	}

	toBalance := root.getOrCreateBalance(record.To, record.Tick)

	fromBalance.SubFreeze(root.Block.Number, record.Amount)
	toBalance.AddAvailable(root.Block.Number, record.Amount)

	return event, nil
}

// ==================== about staking: config pool & stake & unstake & proxy_unstake ====================

func (root *AggregateRoot) getPoolAggregate(pool string) (*staking.PoolAggregate, error) {
	poolRoot, existed := root.StakingPools[pool]
	if !existed {
		return nil, protocol.NewProtocolError(protocol.StakingPoolNotFound, "pool not found")
	}

	return poolRoot, nil
}

func (root *AggregateRoot) handleConfigStaking(command *protocol.ConfigStakeCommand) (err error) {

	ee := &StakingPoolUpdatedEvent{
		BlockNumber:       command.BlockNumber,
		PrevBlockNumber:   root.PreviousBlock,
		TxHash:            command.TxHash,
		PositionInIERCTxs: 0,
		From:              command.From,
		To:                command.To,
		Value:             command.TxValue.String(),
		Data: &StakingPoolUpdated{
			Protocol:  command.Protocol,
			Operate:   command.Operate,
			From:      command.From,
			To:        command.To,
			Pool:      command.Pool,
			PoolID:    command.PoolSubID,
			Name:      command.Name,
			Owner:     command.Owner,
			Admins:    command.Admins,
			Details:   command.Details,
			StopBlock: command.StopBlock,
		},
		ErrCode:   0,
		ErrReason: "",
		EventAt:   command.EventAt,
	}

	defer func() {
		ee.SetError(err)
		root.Events = append(root.Events, ee)
	}()

	for _, record := range command.Details {
		if _, existed := root.TicksMap[record.Tick]; !existed {
			return protocol.NewProtocolError(protocol.StakingTickNotExisted, "invalid tick")
		}
	}

	poolRoot, err := root.getPoolAggregate(command.Pool)
	if err != nil {
		poolRoot = staking.NewPoolAggregate(command.Pool, command.Owner)
		root.StakingPools[poolRoot.PoolAddress] = poolRoot
		err = nil
	}

	return poolRoot.UpdatePool(command)
}

func (root *AggregateRoot) handleStaking(command *protocol.StakingCommand) error {

	poolRoot, err := root.getPoolAggregate(command.Pool)
	if err != nil {
		return err
	}

	if !poolRoot.SubPoolIsExisted(command.PoolSubID) {
		return protocol.NewProtocolError(protocol.StakingPoolNotFound, "pool not found")
	}

	for idx, record := range command.Details {

		ee := &IERC20TransferredEvent{
			BlockNumber:       command.BlockNumber,
			PrevBlockNumber:   root.PreviousBlock,
			TxHash:            command.TxHash,
			PositionInIERCTxs: idx,
			From:              command.From,
			To:                command.To,
			Value:             command.TxValue.String(),
			Data: &IERC20Transferred{
				Protocol: record.Protocol,
				Operate:  record.Operate,
				Tick:     record.Tick,
				From:     record.Staker, // staker
				To:       record.Pool,   // pool
				Amount:   record.Amount,
			},
			ErrCode:   0,
			ErrReason: "",
			EventAt:   command.EventAt,
		}
		root.Events = append(root.Events, ee)

		if err := root.handleStakingRecord(poolRoot, record); err != nil {
			ee.SetError(err)
		}
	}

	return nil
}

func (root *AggregateRoot) handleStakingRecord(pool *staking.PoolAggregate, record *protocol.StakingDetail) error {

	_, existed := root.TicksMap[record.Tick]
	if !existed {
		return protocol.NewProtocolError(protocol.TickNotExist, "tick not exist")
	}

	stakerBalance := root.getOrCreateBalance(record.Staker, record.Tick)

	if record.Amount.GreaterThan(stakerBalance.Available) {
		return protocol.NewProtocolError(
			protocol.InsufficientAvailableFunds,
			fmt.Sprintf("insufficient balance. available(%s) < stake(%s)", stakerBalance.Available, record.Amount),
		)
	}

	err := pool.Staking(root.Block.Number, record.PoolSubID, record.Staker, record.Tick, record.Amount)
	if err != nil {
		return err
	}

	stakerBalance.SubAvailable(root.Block.Number, record.Amount)
	poolBalance := root.getOrCreateBalance(pool.PoolAddress, record.Tick)
	poolBalance.AddFreeze(root.Block.Number, record.Amount)

	return nil
}

func (root *AggregateRoot) handleUnStaking(command *protocol.StakingCommand) error {

	poolRoot, err := root.getPoolAggregate(command.Pool)
	if err != nil {
		return err
	}

	if !poolRoot.SubPoolIsExisted(command.PoolSubID) {
		return protocol.NewProtocolError(protocol.StakingPoolNotFound, "pool not found")
	}

	for idx, record := range command.Details {
		ee := &IERC20TransferredEvent{
			BlockNumber:       command.BlockNumber,
			PrevBlockNumber:   root.PreviousBlock,
			TxHash:            command.TxHash,
			PositionInIERCTxs: idx,
			From:              command.From,
			To:                command.To,
			Value:             command.TxValue.String(),
			Data: &IERC20Transferred{
				Protocol: record.Protocol,
				Operate:  record.Operate,
				Tick:     record.Tick,
				From:     record.Pool,   // pool
				To:       record.Staker, // staker
				Amount:   record.Amount,
			},
			ErrCode:   0,
			ErrReason: "",
			EventAt:   command.EventAt,
		}
		root.Events = append(root.Events, ee)

		if err := root.handleUnStakingRecord(poolRoot, record); err != nil {
			ee.SetError(err)
		}
	}

	return nil
}

func (root *AggregateRoot) handleUnStakingRecord(pool *staking.PoolAggregate, record *protocol.StakingDetail) error {

	_, existed := root.TicksMap[record.Tick]
	if !existed {
		return protocol.NewProtocolError(protocol.TickNotExist, "tick not exist")
	}

	if err := pool.UnStaking(root.Block.Number, record.PoolSubID, record.Staker, record.Tick, record.Amount); err != nil {
		return err
	}

	poolBalance := root.getOrCreateBalance(pool.PoolAddress, record.Tick)
	if poolBalance.Freeze.LessThan(record.Amount) {
		panic("pool freeze funds error, data error")
	}

	poolBalance.SubFreeze(root.Block.Number, record.Amount)

	stakerBalance := root.getOrCreateBalance(record.Staker, record.Tick)
	stakerBalance.AddAvailable(root.Block.Number, record.Amount)

	return nil
}

func (root *AggregateRoot) handleProxyUnStaking(command *protocol.StakingCommand) error {

	poolRoot, err := root.getPoolAggregate(command.Pool)
	if err != nil {
		return err
	}

	if !poolRoot.SubPoolIsExisted(command.PoolSubID) {
		return protocol.NewProtocolError(protocol.StakingPoolNotFound, "pool not found")
	}

	if !poolRoot.IsAdmin(command.PoolSubID, command.From) {
		return protocol.NewProtocolError(protocol.ProxyUnStakingErrNotAdmin, "not admin")
	}

	for idx, record := range command.Details {
		ee := &IERC20TransferredEvent{
			BlockNumber:       command.BlockNumber,
			PrevBlockNumber:   root.PreviousBlock,
			TxHash:            command.TxHash,
			PositionInIERCTxs: idx,
			From:              command.From,
			To:                command.To,
			Value:             command.TxValue.String(),
			Data: &IERC20Transferred{
				Protocol: record.Protocol,
				Operate:  record.Operate,
				Tick:     record.Tick,
				From:     record.Pool,   // pool
				To:       record.Staker, // staker
				Amount:   record.Amount,
			},
			ErrCode:   0,
			ErrReason: "",
			EventAt:   command.EventAt,
		}
		root.Events = append(root.Events, ee)

		err := root.handleUnStakingRecord(poolRoot, record)
		if err != nil {
			ee.SetError(err)
		}
	}

	return nil
}
