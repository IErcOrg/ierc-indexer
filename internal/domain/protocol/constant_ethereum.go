//go:build !sepolia

package protocol

const (
	PlatformAddress = "0x33302dbff493ed81ba2e7e35e2e8e833db023333"

	DPoSMintMintPointsLimitBlockHeight uint64 = 19033750
	DPoSDisableDualMiningBlockHeight   uint64 = 19085665
	PoWMintLimitBlockHeight            uint64 = 19119100
	DPoSMintMinPoints                  int64  = 1000
)
