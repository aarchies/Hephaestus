package consumer

import (
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/aarchies/go-lib/messagec/cqrs"
	"github.com/aarchies/go-lib/messagec/cqrs/message"
	"github.com/sirupsen/logrus"
	"reflect"
)

type IntegrationConsumerGroupHandler[T reflect.Type] struct {
	c *cqrs.IEventBusSubscriptionsManager
}

func NewIntegrationConsumerHandler[T reflect.Type](c *cqrs.IEventBusSubscriptionsManager) *IntegrationConsumerGroupHandler[T] {
	return &IntegrationConsumerGroupHandler[T]{c}
}

func (d *IntegrationConsumerGroupHandler[T]) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (d *IntegrationConsumerGroupHandler[T]) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (d *IntegrationConsumerGroupHandler[T]) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for m := range claim.Messages() {

		var msg message.Message
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			logrus.Errorf("event [%s] messages deserialization error! %s\n", m.Topic, err.Error())
			return err
		}

		//event := new(T)

		//if err := d.h.Handle(context.Background(), msg); err != nil {
		//	return err
		//} else {
		//	session.MarkOffset(m.Topic, m.Partition, m.Offset+1, "")
		//	session.Commit()
		//}
	}

	return nil
}
