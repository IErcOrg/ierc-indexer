package parser

import (
	"strconv"
)

type Uint64 uint64

func (id *Uint64) UnmarshalJSON(data []byte) error {

	strValue, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}

	uintValue, err := strconv.ParseUint(strValue, 10, 64)
	if err != nil {
		return err
	}

	*id = Uint64(uintValue)
	return nil
}
