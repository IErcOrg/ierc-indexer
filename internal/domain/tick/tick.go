package tick

import (
	"github.com/IErcOrg/IERC_Indexer/internal/domain/protocol"
)

type Tick interface {
	GetID() int64
	GetName() string
	GetProtocol() protocol.Protocol
	LastUpdatedBlock() uint64
	Marshal() ([]byte, error)
	Unmarshal([]byte) error
}

var (
	_ Tick = (*IERC20Tick)(nil)
	_ Tick = (*IERCPoWTick)(nil)
)
