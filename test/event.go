package test

import (
	"github.com/aarchies/hephaestus/messagec/cqrs/message"
	"github.com/aarchies/hephaestus/test/pb"
	"github.com/google/uuid"
)

type TestModel struct {
	Data *pb.Weblog
}

func (n *TestModel) GetPayload() interface{} {
	return n.Data
}

func (n *TestModel) Metadata() message.Metadata {
	m := message.Metadata{}
	m.Set("key", "values")
	return m
}

func (n *TestModel) GetId() string {
	return uuid.New().String()
}
