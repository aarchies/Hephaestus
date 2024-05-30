package cqrs

import (
	"github.com/aarchies/hephaestus/messagec/cqrs/event"
)

type Marshaler interface {
	Marshal(e event.IntegrationEvent) ([]byte, error)
	Unmarshal(e *event.IntegrationEvent, v interface{}) (err error)
}
