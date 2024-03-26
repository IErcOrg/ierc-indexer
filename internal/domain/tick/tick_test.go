package tick

import (
	"testing"
	"time"

	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"
)

func TestTick(t *testing.T) {
	suite.Run(t, new(TestTickSuite))
}

type TestTickSuite struct {
	suite.Suite
}

func (s *TestTickSuite) SetupSuite() {

}

func (s *TestTickSuite) TestPoWMint() {

	var entity = IERCPoWTick{
		ID:        0,
		Tick:      "gg",
		Protocol:  "ggg",
		Decimals:  3,
		MaxSupply: decimal.NewFromInt(21_000_000),
		Tokenomics: []protocol.TokenomicsDetail{
			{BlockNumber: 5053000, Amount: decimal.NewFromInt(1000)},
			{BlockNumber: 5153000, Amount: decimal.NewFromInt(500)},
			{BlockNumber: 5253000, Amount: decimal.NewFromInt(250)},
		},
		Rule: protocol.DistributionRule{
			PowRatio:        decimal.NewFromInt(50),
			MinWorkC:        "0x0000",
			DifficultyRatio: decimal.NewFromInt(10),
			PosRatio:        decimal.NewFromInt(50),
			PosPool:         "0x0000000000000000000000000000000000000000",
		},
		Supply:             decimal.Zero,
		PoWAmount:          decimal.Zero,
		PoSAmount:          decimal.Zero,
		LastUpdatedAtBlock: 5053048,
		Creator:            "",
		CreatedAt:          time.Time{},
		UpdatedAt:          time.Time{},
	}

	s.Equal("0", entity.Supply.String(), "supply error")

	pow, pos := entity.Mint(&PoWMintParams{
		BlockNumber:   5053086,
		TotalPoWShare: decimal.Zero,
		MinerPoWShare: decimal.Zero,
		TotalPoSShare: decimal.NewFromInt(500),
		MinerPoSShare: decimal.NewFromInt(500),
	})
	s.Equal("0", pow.String())
	s.Equal("19000", pos.String())
	s.Equal("19000", entity.Supply.String())
	s.Equal(uint64(5053086), entity.LastUpdatedBlock())

	pow, pos = entity.Mint(&PoWMintParams{
		BlockNumber:   5053095,
		TotalPoWShare: decimal.Zero,
		MinerPoWShare: decimal.Zero,
		TotalPoSShare: decimal.NewFromInt(500),
		MinerPoSShare: decimal.NewFromInt(500),
	})
	s.Equal("0", pow.String())
	s.Equal("4500", pos.String())
	s.Equal("23500", entity.Supply.String())
	s.Equal(uint64(5053095), entity.LastUpdatedBlock())

	pow, pos = entity.Mint(&PoWMintParams{
		BlockNumber:   5053166,
		TotalPoWShare: decimal.Zero,
		MinerPoWShare: decimal.Zero,
		TotalPoSShare: decimal.NewFromInt(500),
		MinerPoSShare: decimal.NewFromInt(500),
	})
	s.Equal("0", pow.String())
	s.Equal("35500", pos.String())
	s.Equal("59000", entity.Supply.String())
	s.Equal(uint64(5053166), entity.LastUpdatedBlock())

}
