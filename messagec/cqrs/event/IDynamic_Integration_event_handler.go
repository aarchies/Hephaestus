package event

import (
	"context"
)

type IDynamicIntegrationEventHandler interface {
	Handle(ctx context.Context, data interface{}) error
}
