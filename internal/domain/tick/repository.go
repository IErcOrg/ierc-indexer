package tick

import (
	"context"
)

type TickRepository interface {
	Load(ctx context.Context, name string) (Tick, error)
	Save(ctx context.Context, entities ...Tick) error
}
