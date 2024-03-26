//go:build sepolia
// +build sepolia

package protocol

const (
	PlatformAddress = "0x1878d3363a02f1b5e13ce15287c5c29515000656"

	DPoSMintMintPointsLimitBlockHeight uint64 = 0
	DPoSDisableDualMiningBlockHeight   uint64 = 5152670
	PoWMintLimitBlockHeight            uint64 = 5182950
	DPoSMintMinPoints                  int64  = 1000
)
