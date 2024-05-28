package event

import (
	"github.com/aarchies/go-lib/messagec/cqrs/event"
	"github.com/aarchies/go-lib/messagec/cqrs/message"
	"github.com/google/uuid"
)

type TestModel struct {
	Data string
}

func NewTestModel() event.IntegrationEvent {
	return TestModel{Data: "hello"}
}

func (n TestModel) Metadata() message.Metadata {
	m := message.Metadata{}
	m.Set("key", "values")
	return m
}

func (n TestModel) GetId() string {
	return uuid.New().String()
}
