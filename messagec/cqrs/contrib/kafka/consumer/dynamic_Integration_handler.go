package consumer

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/aarchies/hephaestus/messagec/cqrs"
	"github.com/aarchies/hephaestus/messagec/cqrs/event"
	"github.com/sirupsen/logrus"
)

type DynamicIntegrationConsumerGroupHandler struct {
	h      event.IDynamicIntegrationEventHandler
	config cqrs.EventBusConfig
}

func NewDynamicIntegrationConsumerHandler(h event.IDynamicIntegrationEventHandler, config cqrs.EventBusConfig) *DynamicIntegrationConsumerGroupHandler {
	return &DynamicIntegrationConsumerGroupHandler{h, config}
}

func (d *DynamicIntegrationConsumerGroupHandler) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (d *DynamicIntegrationConsumerGroupHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (d *DynamicIntegrationConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for m := range claim.Messages() {

		var msg event.IntegrationEvent

		if err := d.config.Marshaler.Unmarshal(&msg, m.Value); err != nil {
			logrus.Errorf("event [%s] messages deserialization error! %s\n", m.Topic, err.Error())
			return err
		}

		if err := d.h.Handle(context.Background(), msg.GetPayload()); err != nil {
			return err
		} else {
			session.MarkOffset(m.Topic, m.Partition, m.Offset+1, "")
			session.Commit()
		}
	}

	return nil
}
