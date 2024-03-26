package pkg

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestRecoverToAddress(t *testing.T) {
	message := "galxe.com wants you to sign in with your Ethereum account:\n0x9B95bc06D373857aae5b5B1f1522645344A81513\n\nSign in with Ethereum to the app.\n\nURI: https://galxe.com\nVersion: 1\nChain ID: 1\nNonce: 5HlFcgaHEdJj4TPth\nIssued At: 2023-08-15T12:12:52.310Z\nExpiration Time: 2023-08-22T12:12:52.255Z"
	signature := "0x5b0a9932fb0b922ccfd636462b6f2e0678a22c3d38d44c71da85a3cb8a219aea736c12b3be75a577e0db9446116cff76305cd8da92b6014afd4ad60d2f09bc511c"
	err, address := RecoverToAddress(message, signature)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(address)
	}
}

//{
//"title": "ierc-20 one approve",
//"to": "0x33302dbff493ed81ba2e7e35e2e8e833db023333",
//"tick": "ethi",
//"amt": "7000",
//"value": "0.02793",
//"nonce": "1692357645234"
//}
type TRec struct {
	Title string `json:"title"`
	To    string `json:"to"`
	Tick  string `json:"tick"`
	Amt   string `json:"amt"`
	Value string `json:"value"`
	Nonce string `json:"nonce"`
}
func TestRecoverToAddress1(t *testing.T) {
	ts := TRec{
		Title: "ierc-20 one approve",
		To:    "0x33302dbff493ed81ba2e7e35e2e8e833db023333",
		Tick:  "ethi",
		Amt:   "7000",
		Value: "0.02793",
		Nonce: "1692357645234",
	}
	//s := map[string]interface{}{
	//	"title": "ierc-20 one approve",
	//	"to":    "0x33302dbff493ed81ba2e7e35e2e8e833db023333",
	//	"tick":  "ethi",
	//	"amt":   "7000",
	//	"value": "0.02793",
	//	"nonce": "1692357645234",
	//}

	message, _ := json.MarshalIndent(ts, "", "    ")
	signature := "0x82253ffda5be503c624089280114c1086b9c4aef44032a54dfce5053b06ad90f2946d57102b4652ffe5913919d585c0e5b2eff78c221ec7f66fda3dad29b876d1b"
	err, address := RecoverToAddress(string(message), signature)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(address)
	}
}
