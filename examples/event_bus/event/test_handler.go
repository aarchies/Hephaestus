package event

import (
	"context"
	"encoding/json"
	"github.com/aarchies/go-lib/messagec/cqrs/event"
	"github.com/aarchies/go-lib/messagec/cqrs/message"
	"github.com/sirupsen/logrus"
)

type TestHandler struct{}

var _ event.IDynamicIntegrationEventHandler = TestHandler{}

func (t TestHandler) Handle(ctx context.Context, data message.Message) error {

	var a TestModel
	// json
	json.Unmarshal(data.Payload, &a)

	// protobuf 指定pb模型
	//proto.Unmarshal(data.Payload, &modelPb.Weblog{})

	logrus.Infof("触发Handle %v\n", data)

	return nil
}
