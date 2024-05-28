package event

import (
	"github.com/aarchies/go-lib/messagec/cqrs/message"
)

type IntegrationEvent interface {
	GetId() string
	Metadata() message.Metadata
}
