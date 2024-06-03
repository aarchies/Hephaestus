package cqrs

import (
	"encoding/json"
	"github.com/aarchies/hephaestus/cqrs/event"
	"github.com/aarchies/hephaestus/cqrs/message"
	"reflect"
)

type JsonMarshaler struct{}

func (m JsonMarshaler) Marshal(data interface{}) ([]byte, string, error) {

	e := data.(event.IntegrationEvent)
	b, err := json.Marshal(data)
	if err != nil {
		return nil, "", err
	}

	msg := message.NewMessage(e.GetId(), e.Metadata(), b)
	bytes, err := json.Marshal(msg)
	if err != nil {
		return nil, "", err
	}

	return bytes, msg.UUID, nil
}

func (JsonMarshaler) Unmarshal(e *message.Message, v reflect.Value) (err error) {
	return json.Unmarshal(e.Payload, v.Interface())
}
