package event

import "github.com/aarchies/hephaestus/cqrs/message"

type IntegrationEventHandler[T any] interface {
	Handle(uid string, metadata message.Metadata, data T) error
}
