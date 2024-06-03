package event

import "github.com/aarchies/hephaestus/messagec/cqrs/message"

type IntegrationEventHandler[T any] interface {
	Handle(uid string, metadata message.Metadata, data T) error
}
