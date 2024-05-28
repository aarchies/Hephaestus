package event

import (
	"context"
	"flow_crafter_CDN/pkg/messagec/cqrs/message"
)

type IDynamicIntegrationEventHandler interface {
	Handle(ctx context.Context, data message.Message) error
}
