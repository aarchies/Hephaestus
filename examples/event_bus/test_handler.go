package event_bus

import (
	"github.com/aarchies/hephaestus/messagec/cqrs/event"
	"github.com/aarchies/hephaestus/messagec/cqrs/message"
	"github.com/sirupsen/logrus"
)

type TestHandler struct{}

var _ event.IntegrationEventHandler = TestHandler{}

func (t TestHandler) Handle(uid string, metadata message.Metadata, data interface{}) error {

	// json

	// protobuf 指定pb模型
	//proto.Unmarshal(data.Payload, &modelPb.Weblog{})

	logrus.Infof("触发Handle方法! eventId:%s metaData:%v data:%v\n", uid, metadata, data.(TestModel))

	return nil
}
