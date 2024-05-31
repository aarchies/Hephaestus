package test

import (
	"github.com/aarchies/hephaestus/messagec/cqrs/message"
	"github.com/google/uuid"
)

type EventModel struct {
	//Data *pb.Weblog
}

//func (n *EventModel) GetPayload() interface{} {
//	return n.Data
//}

func (n EventModel) Metadata() message.Metadata {
	m := message.Metadata{}
	m.Set("key", "values")
	return m
}

func (n EventModel) GetId() string {
	return uuid.New().String()
}
