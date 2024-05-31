package cqrs

import (
	"encoding/json"
	"github.com/aarchies/hephaestus/messagec/cqrs/event"
	"github.com/aarchies/hephaestus/messagec/cqrs/message"

	"reflect"
)

type JsonMarshaler struct{}

func (m JsonMarshaler) Marshal(e event.IntegrationEvent) ([]byte, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}

	msg := message.NewMessage(e.GetId(), e.Metadata(), b)
	msg.Metadata.Set("name", reflect.TypeOf(e).Elem().Name())

	bytes, err := json.Marshal(message.NewMessage(e.GetId(), e.Metadata(), b))
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (JsonMarshaler) Unmarshal(e *message.Message, v interface{}) (err error) {
	//return json.Unmarshal(msg.Payload, v)
	return nil
}
