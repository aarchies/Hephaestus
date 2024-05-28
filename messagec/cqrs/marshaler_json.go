package cqrs

import (
	"encoding/json"
	"flow_crafter_CDN/pkg/messagec/cqrs/event"
	"flow_crafter_CDN/pkg/messagec/cqrs/message"
	"reflect"
)

type JsonMarshaler struct{}

func (m JsonMarshaler) Marshal(v interface{}) (*message.Message, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	e := v.(event.IntegrationEvent)

	msg := message.NewMessage(e.GetId(), e.Metadata(), b)

	msg.Metadata.Set("name", reflect.TypeOf(v).Name())

	return msg, nil
}

func (JsonMarshaler) Unmarshal(msg *message.Message, v interface{}) (err error) {
	return json.Unmarshal(msg.Payload, v)
}
