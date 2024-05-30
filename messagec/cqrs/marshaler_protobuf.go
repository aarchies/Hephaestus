package cqrs

import (
	"encoding/json"
	"fmt"
	"github.com/aarchies/hephaestus/messagec/cqrs/message"
	"github.com/aarchies/hephaestus/messagec/cqrs/message/pb"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"reflect"
)

type ProtobufMarshaler struct{}

func (m ProtobufMarshaler) Marshal(v interface{}) ([]byte, error) {

	b, err := proto.Marshal(v.(proto.Message))
	if err != nil {
		err := fmt.Sprintf("protobuf序列化消息时发生错误! Event:%s %v", reflect.TypeOf(v).Name(), v)
		return nil, errors.New(err)
	}

	return b, nil
}

func (ProtobufMarshaler) Unmarshal(msg *message.Message, v interface{}) (err error) {
	m := &pb.Message{}
	if err := proto.Unmarshal(v.([]byte), m); err != nil {
		return errors.New(fmt.Sprintf("protobuf反序列化消息时发生错误! Event:%s %v", reflect.TypeOf(v).Name(), v))
	}
	msg.UUID = m.Uid
	msg.Metadata = m.MateData
	msg.Time = m.Time.AsTime()

	if err := json.Unmarshal(m.Payload, &msg.Payload); err != nil {
		return err
	}

	return nil
}
