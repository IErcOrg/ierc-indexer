package protocol

import (
	"fmt"
)

type ProtocolErrCode int

const (
	ProtocolErr ProtocolErrCode = iota + 0x0100
	NotProtocolData
	InvalidProtocolFormat
	InvalidProtocolParams
	UnknownProtocol
	UnknownProtocolOperate
	InvalidTxHash

	TickNotExist
	TickExited

	InsufficientAvailableFunds
	InsufficientFreezeFunds
	InsufficientValue

	SignatureNotExist
	SignatureAlreadyUsed
	SignatureNotMatch

	MintErr ProtocolErrCode = iota + 0x0200
	MintErrTickNotFound
	MintErrTickNotSupportPoW
	MintErrTickProtocolNoMatch
	MintErrTickMinted
	MintPoWInvalidHash
	MintPoSInvalidShare
	MintAlreadyMinted
	MintAmountExceedLimit
	MintInvalidBlock
	MintBlockExpires
	MintErrMaxAmountLessThanSupply
	MintErrNoPermissionToClaimAirdrop
	MintErrInvalidAirdropAmount
	MintErrAirdropAmountExceedsRemainSupply
	MintErrAirdropClaimFailed

	InvalidSignature

	TickErr ProtocolErrCode = iota + 0x0300
	ErrTickProtocolNoMatch
	ErrUpdateMaxSupplyNoPermission
	ErrUpdateAmountLessThanSupply
	ErrUpdateFailed

	UnknownError ProtocolErrCode = iota + 0x0800

	StakingError ProtocolErrCode = iota + 0x0900
	StakingTickUnsupported
	StakingTickNotExisted
	StakingPoolNotFound
	StakingPoolAlreadyStopped
	StakingPoolIsFulled
	StakingPoolIsEnded
	StakingPoolMaxAmountLessThanCurrentAmount
	StakeConfigPoolNotMatch
	StakeConfigNoPermission
	UnStakingErrNoStake
	UnStakingErrStakeAmountInsufficient
	UnStakingErrNotYetUnlocked
	ProxyUnStakingErrNotAdmin
	UseRewardsErrNoStake
	UseRewardsErrRewardsInsufficient
	MintErrDPoSMintPointsTooLow
	MintErrPoWShareZero
)

type ProtocolError struct {
	code    ProtocolErrCode
	message string
}

func (e *ProtocolError) Code() int32 { return int32(e.code) }

func (e *ProtocolError) Message() string { return e.message }

func (e *ProtocolError) Error() string {
	return fmt.Sprintf("error code: %d, message: %s", e.code, e.message)
}

func NewProtocolError(code ProtocolErrCode, message string) *ProtocolError {
	return &ProtocolError{
		code:    code,
		message: message,
	}
}
