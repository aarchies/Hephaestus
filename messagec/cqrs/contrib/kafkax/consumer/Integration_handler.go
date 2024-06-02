package consumer

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/aarchies/hephaestus/messagec/cqrs"
	"github.com/aarchies/hephaestus/messagec/cqrs/message"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"reflect"
)

type IntegrationConsumerGroupHandler struct {
	subscriptionsManager *cqrs.IEventBusSubscriptionsManager
	config               cqrs.EventBusConfig
}

func NewIntegrationConsumerHandler(subscriptionsManager *cqrs.IEventBusSubscriptionsManager, config cqrs.EventBusConfig) *IntegrationConsumerGroupHandler {
	return &IntegrationConsumerGroupHandler{subscriptionsManager, config}
}

func (d IntegrationConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (d IntegrationConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (d IntegrationConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for m := range claim.Messages() {

		var msg message.Message
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			logrus.Errorf("event [%s] messages deserialization error! %s\n", m.Topic, err.Error())
			return err
		}

		// 创建事件模型实例
		model := reflect.New(d.subscriptionsManager.GetEventType(m.Topic))
		// 反序列化消息到事件模型实例
		if err := d.config.Marshaler.Unmarshal(&msg, model.Interface()); err != nil {
			logrus.Errorf("Parsing the event model failed! err:%s\n", err.Error())
			return err
		}

		// 创建处理程序实例
		handler := reflect.New(d.subscriptionsManager.GetHandler(m.Topic).HandlerType.Elem()).Interface()
		// 调用处理程序的 Handle 方法
		method := reflect.ValueOf(handler).MethodByName("Handle")

		if !method.IsValid() {
			return errors.New("The Handle method was not found!")
		}

		method.Call([]reflect.Value{reflect.ValueOf(msg.UUID), reflect.ValueOf(msg.Metadata), reflect.ValueOf(model.Elem().Elem().Interface())})

		session.MarkOffset(m.Topic, m.Partition, m.Offset+1, "")
		session.Commit()
	}

	return nil
}
