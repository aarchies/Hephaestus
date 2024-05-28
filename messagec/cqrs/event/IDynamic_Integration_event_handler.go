package event

import (
	"context"
	"github.com/aarchies/go-lib/messagec/cqrs/message"
)

type IDynamicIntegrationEventHandler interface {
	Handle(ctx context.Context, data message.Message) error
}
