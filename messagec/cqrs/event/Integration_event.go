package event

import (
	"flow_crafter_CDN/pkg/messagec/cqrs/message"
)

type IntegrationEvent interface {
	GetId() string
	Metadata() message.Metadata
}
