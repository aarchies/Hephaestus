package cqrs

import (
	"github.com/aarchies/hephaestus/messagec/cqrs/event"
	"github.com/aarchies/hephaestus/messagec/cqrs/message"
)

type Marshaler interface {
	Marshal(e event.IntegrationEvent) ([]byte, error)
	Unmarshal(e *message.Message, v interface{}) (err error)
}
