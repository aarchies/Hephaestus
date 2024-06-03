package event_bus

import (
	"github.com/aarchies/hephaestus/cqrs/message"
	"github.com/aarchies/hephaestus/examples/event_bus/pb"
	"github.com/google/uuid"
)

type TestModel struct {
	*pb.Weblog
	//Str string
}

func (n TestModel) Metadata() message.Metadata {
	m := message.Metadata{}
	m.Set("key", "values")
	return m
}

func (n TestModel) GetId() string {
	return uuid.New().String()
}
