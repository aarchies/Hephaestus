package cqrs

import "context"

type EventHandler interface {
	GetHandlerName() string
	Handle(ctx context.Context, cmd any) error
}
