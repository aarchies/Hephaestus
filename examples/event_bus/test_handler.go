package event_bus

import (
	"github.com/aarchies/hephaestus/cqrs/event"
	"github.com/aarchies/hephaestus/cqrs/message"
	"github.com/aarchies/hephaestus/examples/event_bus/pb"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
)

type TestHandler struct{}

var _ event.IntegrationEventHandler[TestModel] = TestHandler{}

func (t TestHandler) Handle(uid string, metadata message.Metadata, data TestModel) error {

	logrus.Infof("触发Handle方法! eventId:%s metaData:%v data:%v\n", uid, metadata, data)

	return nil
}

type TestDynamicHandler struct{}

var _ event.IDynamicIntegrationEventHandler = TestDynamicHandler{}

func (t TestDynamicHandler) Handle(uid string, metadata message.Metadata, data interface{}) error {
	model := TestModel{
		Weblog: &pb.Weblog{},
	}

	if err := proto.Unmarshal(data.([]byte), &model); err != nil {
		return err
	}
	logrus.Infof("触发Dynamic Handle方法! eventId:%s metaData:%v data:%v\n", uid, metadata, string(data.([]byte)))

	return nil
}
