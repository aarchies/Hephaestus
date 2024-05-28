package cqrs

import (
	"fmt"
	"github.com/aarchies/go-lib/messagec/cqrs/event"
	"github.com/aarchies/go-lib/messagec/cqrs/message"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"reflect"
)

type ProtobufMarshaler struct{}

func (m ProtobufMarshaler) Marshal(v interface{}) (*message.Message, error) {
	protoMsg, ok := v.(proto.Message)
	if !ok {
		err := fmt.Sprintf("序列化消息时发生错误! Event:%s %v", reflect.TypeOf(v).Name(), v)
		return nil, errors.New(err)
	}

	b, err := proto.Marshal(protoMsg)
	if err != nil {
		return nil, err
	}

	e := v.(event.IntegrationEvent)
	msg := message.NewMessage(e.GetId(), e.Metadata(), b)

	msg.Metadata.Set("name", reflect.TypeOf(v).Name())

	return msg, nil
}

func (ProtobufMarshaler) Unmarshal(msg *message.Message, v interface{}) (err error) {
	return proto.Unmarshal(msg.Payload, v.(proto.Message))
}
