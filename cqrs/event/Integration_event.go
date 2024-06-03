package event

import (
	"github.com/aarchies/hephaestus/cqrs/message"
)

type IntegrationEvent interface {
	GetId() string
	Metadata() message.Metadata
}
