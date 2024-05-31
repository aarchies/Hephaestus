package cqrs

import (
	"encoding/json"
	"fmt"
	"github.com/aarchies/hephaestus/messagec/cqrs/event"
	"github.com/aarchies/hephaestus/messagec/cqrs/message"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"reflect"
)

type ProtobufMarshaler struct{}

func (m ProtobufMarshaler) Marshal(e event.IntegrationEvent) ([]byte, error) {

	b, err := proto.Marshal(e.(proto.Message))
	if err != nil {
		err := fmt.Sprintf("protobuf序列化消息时发生错误! Event:%s %s", err.Error(), reflect.TypeOf(e).Elem().Name())
		return nil, errors.New(err)
	}
	e.Metadata().Set("name", reflect.TypeOf(e).Elem().Name())
	bytes, err := json.Marshal(message.NewMessage(e.GetId(), e.Metadata(), b))
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (ProtobufMarshaler) Unmarshal(e *message.Message, v interface{}) (err error) {

	if err := json.Unmarshal(v.([]byte), &e); err != nil {
		return errors.New(fmt.Sprintf("protobuf反序列化消息时发生错误! Event:%s", err.Error()))
	}

	return nil
}
