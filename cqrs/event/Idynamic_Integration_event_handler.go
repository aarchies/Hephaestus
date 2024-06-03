package event

import "github.com/aarchies/hephaestus/cqrs/message"

type IDynamicIntegrationEventHandler interface {
	Handle(uid string, metadata message.Metadata, data interface{}) error
}
