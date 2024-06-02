package event

import "github.com/aarchies/hephaestus/messagec/cqrs/message"

type IntegrationEventHandler interface {
	Handle(uid string, metadata message.Metadata, data interface{}) error
}
