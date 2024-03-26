package protocol

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestSignature(t *testing.T) {
	suite.Run(t, new(TestSignatureSuite))
}

type TestSignatureSuite struct {
	suite.Suite
}

func (s *TestSignatureSuite) SetupSuite() {

}

func (s *TestSignatureSuite) TestSignature() {
	sign := "0x00052a3c417bc511cbb71890e5023eb32533a8083d3d23de1838f1e0fca944bd25a86476f32415ade361ad616450264f0aa874c2f5b6e8aceb2bde0313112b8c1b"
	signature := NewSignature(
		"ierc-m4",
		"0x7ca8a0a62a61af7ccd440649232d6a79d26434ac",
		"0x33302dbff493ed81ba2e7e35e2e8e833db023333",
		"5000",
		"0.045",
		"1700802840255",
	)
	err := signature.ValidSignature(sign)
	s.Nil(err)
}

func (s *TestSignatureSuite) TestSignature1() {
	sign := "0x69e86aa9f792aa0b8a146fc3b2946ee33fc76cf7f1fe0736895f5e4a72eea1a661dd742590913b724f10ce41bca9d663f55653098daea13437f616042d2e56e31c"
	signature := NewSignature(
		"ethi",
		"0x9ffc341849486014b340f8d7a3fad10e972aede6",
		"0x1878d3363a02f1b5e13ce15287c5c29515000656",
		"1",
		"0.005",
		"1703841847886",
	)
	err := signature.ValidSignature(sign)
	s.Nil(err)
}
