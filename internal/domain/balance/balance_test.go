package balance

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/suite"
)

func TestBalance(t *testing.T) {
	suite.Run(t, new(TestBalanceSuite))
}

type TestBalanceSuite struct {
	suite.Suite
}

func (s *TestBalanceSuite) SetupSuite() {

}

func (s *TestBalanceSuite) TestBalanceKey() {
	key1 := NewBalanceKey("0x11", "ethi")
	key2 := NewBalanceKey("0x11", "ethi")

	var valueMap = make(map[BalanceKey]struct{})
	valueMap[key1] = struct{}{}
	valueMap[key2] = struct{}{}
	s.Equal(1, len(valueMap))

	var ptrMap = make(map[*BalanceKey]struct{})
	ptrMap[&key1] = struct{}{}
	ptrMap[&key2] = struct{}{}
	s.Equal(2, len(ptrMap))
	for key, _ := range ptrMap {
		spew.Dump(key, fmt.Sprintf("%p", key))
	}
}
