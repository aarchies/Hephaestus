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

func (m ProtobufMarshaler) Marshal(data interface{}) ([]byte, string, error) {

	e := data.(event.IntegrationEvent)
	b, err := proto.Marshal(data.(proto.Message))
	if err != nil {
		err := fmt.Sprintf("protobuf序列化消息时发生错误! Event:%s %s", err.Error(), reflect.TypeOf(e).Elem().Name())
		return nil, "", errors.New(err)
	}

	msg := message.NewMessage(e.GetId(), e.Metadata(), b)
	bytes, err := json.Marshal(msg)
	if err != nil {
		return nil, "", err
	}
	return bytes, msg.UUID, nil
}

func (ProtobufMarshaler) Unmarshal(e *message.Message, v reflect.Value) (err error) {

	// 初始化成员类型
	v.Elem().Set(reflect.New(reflect.TypeOf(v.Elem().Interface()).Elem()))

	// 初始化成员值
	valueField := reflect.ValueOf(v.Elem().Interface()).Elem()
	for i := 0; i < valueField.NumField(); i++ {
		field := valueField.Field(i)
		if field.Kind() == reflect.Ptr {
			field.Set(reflect.New(field.Type().Elem()))
		}
	}

	return proto.Unmarshal(e.Payload, v.Elem().Interface().(proto.Message))
}
