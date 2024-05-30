package event

import (
	"github.com/aarchies/hephaestus/messagec/cqrs/message"
)

type IntegrationEvent interface {
	GetId() string
	Metadata() message.Metadata
	GetPayload() interface{}
}
