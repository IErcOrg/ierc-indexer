package protocol

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

type Signature struct {
	Title  string `json:"title"`
	Signer string `json:"-"`
	To     string `json:"to"`
	Tick   string `json:"tick"`
	Amt    string `json:"amt"`
	Value  string `json:"value"`
	Nonce  string `json:"nonce"`
}

func NewSignature(tick, signer, to, amount, value, nonce string) *Signature {
	return &Signature{
		Title:  SignatureTitle,
		Signer: signer,
		To:     to,
		Tick:   tick,
		Amt:    amount,
		Value:  value,
		Nonce:  nonce,
	}
}

func (s *Signature) ValidSignature(signature string) error {
	if len(signature) == 0 || !strings.HasPrefix(strings.ToLower(signature), "0x") {
		return NewProtocolError(InvalidSignature, "invalid sign format")
	}
	sig, err := hex.DecodeString(strings.TrimPrefix(signature, "0x"))
	if err != nil {
		return NewProtocolError(InvalidSignature, "ValidateEOASignature, signature is an invalid hex string")
	}
	if len(sig) != 65 {
		return NewProtocolError(InvalidSignature, "ValidateEOASignature, signature is not of proper length")
	}
	if sig[64] > 1 {
		sig[64] -= 27 // recovery ID
	}

	message, _ := json.MarshalIndent(s, "", "    ")
	hash := crypto.Keccak256([]byte(fmt.Sprintf("\x19Ethereum Signed Message:\n%v%s", len(message), message)))

	pubKey, err := crypto.SigToPub(hash, sig)
	if err != nil {
		return NewProtocolError(InvalidSignature, err.Error())
	}

	key := crypto.PubkeyToAddress(*pubKey)
	if strings.ToLower(key.Hex()) != s.Signer {
		return NewProtocolError(SignatureNotMatch, "signature not match")
	}

	return nil
}
