package handler

import (
	pb "github.com/IErcOrg/IERC_Indexer/api/indexer"
	"github.com/IErcOrg/IERC_Indexer/internal/domain"
	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
)

func ConvertEventEntityToProtobuf(item domain.Event) *pb.Event {

	switch ee := item.(type) {
	case *domain.IERC20TickCreatedEvent:
		return convertTickCreatedToPB(ee)

	case *domain.IERC20MintedEvent:
		return convertMintedToPB(ee)

	case *domain.IERCPoWTickCreatedEvent:
		return convertPowTickCreatedToPB(ee)

	case *domain.IERCPoWMintedEvent:
		return convertPowMintedToPB(ee)

	case *domain.IERC20TransferredEvent:
		return convertTickTransferredEventToPB(ee)

	case *domain.StakingPoolUpdatedEvent:
		return convertStakingPoolUpdatedToPB(ee)

	default:
		panic("invalid event type")
	}
}

var operateMap = map[protocol.Operate]pb.Operate{
	protocol.OpDeploy:          pb.Operate_Deploy,
	protocol.OpMint:            pb.Operate_Mint,
	protocol.OpTransfer:        pb.Operate_Transfer,
	protocol.OpFreezeSell:      pb.Operate_FreezeSell,
	protocol.OpUnfreezeSell:    pb.Operate_UnfreezeSell,
	protocol.OpProxyTransfer:   pb.Operate_ProxyTransfer,
	protocol.OpStakeConfig:     pb.Operate_StakeConfig,
	protocol.OpStaking:         pb.Operate_Stake,
	protocol.OpUnStaking:       pb.Operate_UnStake,
	protocol.OpProxyUnStaking:  pb.Operate_ProxyUnStake,
	protocol.OpPoWModify:       pb.Operate_Modify,
	protocol.OpPoWClaimAirdrop: pb.Operate_ClaimAirdrop,
}

func convertOperate(operate protocol.Operate) pb.Operate {
	op, existed := operateMap[operate]
	if !existed {
		return pb.Operate_OPERATE_UNSPECIFIED
	}
	return op
}

func convertTickCreatedToPB(ee *domain.IERC20TickCreatedEvent) *pb.Event {
	return &pb.Event{
		BlockNumber:  ee.BlockNumber,
		TxHash:       ee.TxHash,
		PosInIercTxs: int32(ee.PositionInIERCTxs),
		From:         ee.From,
		To:           ee.To,
		Value:        ee.Value,
		EventAt:      ee.EventAt.UnixMilli(),
		ErrCode:      ee.ErrCode,
		ErrReason:    ee.ErrReason,
		Event: &pb.Event_TickCreated{TickCreated: &pb.IERC20TickCreated{
			Protocol:    string(ee.Data.Protocol),
			Operate:     convertOperate(ee.Data.Operate),
			Tick:        ee.Data.Tick,
			Decimals:    ee.Data.Decimals,
			MaxSupply:   ee.Data.MaxSupply.String(),
			Limit:       ee.Data.Limit.String(),
			WalletLimit: ee.Data.WalletLimit.String(),
			Workc:       ee.Data.WorkC,
			Creator:     ee.From,
			Nonce:       ee.Data.Nonce,
		}},
	}
}

func convertMintedToPB(ee *domain.IERC20MintedEvent) *pb.Event {
	return &pb.Event{
		BlockNumber:  ee.BlockNumber,
		TxHash:       ee.TxHash,
		PosInIercTxs: int32(ee.PositionInIERCTxs),
		From:         ee.From,
		To:           ee.To,
		Value:        ee.Value,
		EventAt:      ee.EventAt.UnixMilli(),
		ErrCode:      ee.ErrCode,
		ErrReason:    ee.ErrReason,
		Event: &pb.Event_Minted{Minted: &pb.IERC20Minted{
			Protocol:     string(ee.Data.Protocol),
			Operate:      convertOperate(ee.Data.Operate),
			Tick:         ee.Data.Tick,
			From:         ee.Data.From,
			To:           ee.Data.To,
			Nonce:        ee.Data.Nonce,
			MintedAmount: ee.Data.MintedAmount.String(),
			Gas:          ee.Data.Gas.String(),
			GasPrice:     ee.Data.GasPrice.String(),
		}},
	}
}

func convertPowTickCreatedToPB(ee *domain.IERCPoWTickCreatedEvent) *pb.Event {
	return &pb.Event{
		BlockNumber:  ee.BlockNumber,
		TxHash:       ee.TxHash,
		PosInIercTxs: int32(ee.PositionInIERCTxs),
		From:         ee.From,
		To:           ee.To,
		Value:        ee.Value,
		EventAt:      ee.EventAt.UnixMilli(),
		ErrCode:      ee.ErrCode,
		ErrReason:    ee.ErrReason,
		Event: &pb.Event_PowTickCreated{PowTickCreated: &pb.IERCPoWTickCreated{
			Protocol:          string(ee.Data.Protocol),
			Operate:           convertOperate(ee.Data.Operate),
			Tick:              ee.Data.Tick,
			Decimals:          ee.Data.Decimals,
			MaxSupply:         ee.Data.MaxSupply.String(),
			TokenomicsDetails: convertTokenomicsDetails(ee.Data.Tokenomics),
			Rule:              convertDistributionRuleToPB(&ee.Data.Rule),
			Creator:           ee.From,
		}},
	}
}

func convertTokenomicsDetails(details []protocol.TokenomicsDetail) []*pb.IERCPoWTickCreated_TokenomicsDetail {
	var result = make([]*pb.IERCPoWTickCreated_TokenomicsDetail, 0, len(details))
	for _, detail := range details {
		result = append(result, &pb.IERCPoWTickCreated_TokenomicsDetail{
			BlockNumber: detail.BlockNumber,
			Amount:      detail.Amount.String(),
		})
	}
	return result
}

func convertDistributionRuleToPB(rule *protocol.DistributionRule) *pb.IERCPoWTickCreated_Rule {
	return &pb.IERCPoWTickCreated_Rule{
		PowRatio:        rule.PowRatio.String(),
		MinWorkc:        rule.MinWorkC,
		DifficultyRatio: rule.DifficultyRatio.String(),
		PosRatio:        rule.PosRatio.String(),
		PosPool:         rule.PosPool,
	}
}

func convertPowMintedToPB(ee *domain.IERCPoWMintedEvent) *pb.Event {
	return &pb.Event{
		BlockNumber:  ee.BlockNumber,
		TxHash:       ee.TxHash,
		PosInIercTxs: int32(ee.PositionInIERCTxs),
		From:         ee.From,
		To:           ee.To,
		Value:        ee.Value,
		EventAt:      ee.EventAt.UnixMilli(),
		ErrCode:      ee.ErrCode,
		ErrReason:    ee.ErrReason,
		Event: &pb.Event_PowMinted{PowMinted: &pb.IERCPoWMinted{
			Protocol:        string(ee.Data.Protocol),
			Operate:         convertOperate(ee.Data.Operate),
			Tick:            ee.Data.Tick,
			From:            ee.Data.From,
			To:              ee.Data.To,
			Nonce:           ee.Data.Nonce,
			IsPow:           ee.Data.IsPoW,
			PowTotalShare:   ee.Data.PoWTotalShare.String(),
			PowMinerShare:   ee.Data.PoWMinerShare.String(),
			PowMintedAmount: ee.Data.PoWMintedAmount.String(),
			IsPos:           ee.Data.IsPoS,
			PosTotalShare:   ee.Data.PoSTotalShare.String(),
			PosMinerShare:   ee.Data.PoSMinerShare.String(),
			PosMintedAmount: ee.Data.PoSMintedAmount.String(),
			Gas:             ee.Data.Gas.String(),
			GasPrice:        ee.Data.GasPrice.String(),
			IsAirdrop:       ee.Data.IsAirdrop,
			AirdropAmount:   ee.Data.AirdropAmount.String(),
			BurnedAmount:    ee.Data.BurnAmount.String(),
		}},
	}
}

func convertTickTransferredEventToPB(ee *domain.IERC20TransferredEvent) *pb.Event {
	return &pb.Event{
		BlockNumber:  ee.BlockNumber,
		TxHash:       ee.TxHash,
		PosInIercTxs: int32(ee.PositionInIERCTxs),
		From:         ee.From,
		To:           ee.To,
		Value:        ee.Value,
		EventAt:      ee.EventAt.UnixMilli(),
		ErrCode:      ee.ErrCode,
		ErrReason:    ee.ErrReason,
		Event: &pb.Event_TickTransferred{TickTransferred: &pb.TickTransferred{
			Protocol:    string(ee.Data.Protocol),
			Operate:     convertOperate(ee.Data.Operate),
			Tick:        ee.Data.Tick,
			From:        ee.Data.From,
			To:          ee.Data.To,
			Amount:      ee.Data.Amount.String(),
			EthValue:    ee.Data.EthValue.String(),
			GasPrice:    ee.Data.GasPrice.String(),
			SignerNonce: ee.Data.SignerNonce,
			Sign:        ee.Data.Sign,
		}},
	}
}

func convertStakingPoolUpdatedToPB(ee *domain.StakingPoolUpdatedEvent) *pb.Event {
	return &pb.Event{
		BlockNumber:  ee.BlockNumber,
		TxHash:       ee.TxHash,
		PosInIercTxs: int32(ee.PositionInIERCTxs),
		From:         ee.From,
		To:           ee.To,
		Value:        ee.Value,
		EventAt:      ee.EventAt.UnixMilli(),
		ErrCode:      ee.ErrCode,
		ErrReason:    ee.ErrReason,
		Event: &pb.Event_PoolUpdated{PoolUpdated: &pb.StakingPoolUpdated{
			Protocol:  string(ee.Data.Protocol),
			Operate:   convertOperate(ee.Data.Operate),
			From:      ee.From,
			To:        ee.To,
			Pool:      ee.Data.Pool,
			PoolId:    ee.Data.PoolID,
			Name:      ee.Data.Name,
			Owner:     ee.Data.Owner,
			Admins:    ee.Data.Admins,
			Details:   convertTickConfigDetails(ee.Data.Details),
			StopBlock: ee.Data.StopBlock,
		}},
	}
}

func convertTickConfigDetails(details []*protocol.TickConfigDetail) []*pb.StakingPoolUpdated_TickConfigDetail {
	var result = make([]*pb.StakingPoolUpdated_TickConfigDetail, 0, len(details))
	for _, detail := range details {
		result = append(result, &pb.StakingPoolUpdated_TickConfigDetail{
			Tick:      detail.Tick,
			Ratio:     detail.RewardsRatioPerBlock.String(),
			MaxAmount: detail.MaxAmount.String(),
		})
	}
	return result

}
