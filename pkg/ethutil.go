package pkg

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	privkey  = ""
	endpoint = "https://mainnet.infura.io/v3/d49aedc5c8d04128ab366779756cfacd"
)

type Address common.Address

func RecoverToAddress(message, signature string) (error, string) {
	sig, err := hex.DecodeString(strings.TrimPrefix(signature, "0x"))
	if err != nil {
		return errors.New("ValidateEOASignature, signature is an invalid hex string"), ""
	}
	if len(sig) != 65 {
		return errors.New("ValidateEOASignature, signature is not of proper length"), ""
	}
	msg := fmt.Sprintf("\x19Ethereum Signed Message:\n%v%s", len(message), message)
	hash := crypto.Keccak256([]byte(msg))
	if sig[64] > 1 {
		sig[64] -= 27 // recovery ID
	}
	pubKey, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return err, ""
	}
	key := Address(crypto.PubkeyToAddress(*pubKey))

	return nil, strings.ToLower(common.Address(key).Hex())
}

func GetTxSender(tx *types.Transaction) (common.Address, error) {
	var signer types.Signer
	switch {
	case tx.Type() == types.AccessListTxType:
		signer = types.NewEIP2930Signer(tx.ChainId())
	case tx.Type() == types.DynamicFeeTxType:
		signer = types.NewLondonSigner(tx.ChainId())
	default:
		signer = types.NewEIP155Signer(tx.ChainId())
	}
	sender, err := types.Sender(signer, tx)
	return sender, err
}
