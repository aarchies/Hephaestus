package event_bus

import (
	"github.com/aarchies/hephaestus/messagec/cqrs/event"
	"github.com/aarchies/hephaestus/messagec/cqrs/message"
	"github.com/sirupsen/logrus"
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
	logrus.Infof("触发Dynamic Handle方法! eventId:%s metaData:%v data:%v\n", uid, metadata, data)

	return nil
}
