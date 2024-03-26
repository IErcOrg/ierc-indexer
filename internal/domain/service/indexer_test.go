package service

import (
	"encoding/json"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/shopspring/decimal"
)

func TestWithRetryCount(t *testing.T) {

	//r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//
	//var fn = func() error {
	//	time.Sleep(time.Second * time.Duration(r.Intn(5)+1))
	//	return errors.New("----")
	//}
	//
	//err := WithRetryCount(5, time.Second*1, time.Second*5, fn)()
	//if err != nil {
	//	spew.Dump(err.Error())
	//}

}

func TestName(t *testing.T) {
	v := `{"value": "0.11"}`

	d := struct {
		Value  decimal.Decimal `json:"value"`
		Value1 decimal.Decimal `json:"value1"`
	}{}

	err := json.Unmarshal([]byte(v), &d)
	if err != nil {
		t.Fatal(err)
	}
	spew.Dump(d)
}
