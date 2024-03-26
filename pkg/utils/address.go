package utils

import (
	"github.com/ethereum/go-ethereum/common"
)

func has0xPrefix(str string) bool {
	return len(str) >= 2 && str[0] == '0' && (str[1] == 'x' || str[1] == 'X')
}

func IsHexAddressWith0xPrefix(s string) bool {
	return has0xPrefix(s) && common.IsHexAddress(s)
}
