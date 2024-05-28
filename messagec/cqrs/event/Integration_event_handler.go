package event

import (
	"context"
)

type IntegrationEventHandler[T any] interface {
	Handle(ctx context.Context, event T) error
}
