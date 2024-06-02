package event_bus

import (
	"github.com/aarchies/hephaestus/messagec/cqrs/message"
	"github.com/google/uuid"
)

type TestModel struct {
	Data string
}

func (n TestModel) Metadata() message.Metadata {
	m := message.Metadata{}
	m.Set("key", "values")
	return m
}

func (n TestModel) GetId() string {
	return uuid.New().String()
}
