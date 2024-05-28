package kafka

import (
	"context"
	"encoding/json"
	"flow_crafter_CDN/pkg/messagec/cqrs"
	"flow_crafter_CDN/pkg/messagec/cqrs/contrib/kafka/consumer"
	"flow_crafter_CDN/pkg/messagec/cqrs/event"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/avast/retry-go"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

type BusKafka struct {
	connection    IDefaultKafkaConnection
	asyncProducer sarama.AsyncProducer
	*cqrs.IEventBusSubscriptionsManager
	config cqrs.EventBusConfig
}

func NewEventBusWithConfig(connection IDefaultKafkaConnection, config cqrs.EventBusConfig) cqrs.IEventBus {
	producer, err := sarama.NewAsyncProducerFromClient(connection.GetClient())
	if err != nil {
		logrus.Fatalln("creating Producer Error! %s", err.Error())
	}

	return BusKafka{
		connection:                    connection,
		asyncProducer:                 producer,
		IEventBusSubscriptionsManager: cqrs.SubscriptionsManager,
		config:                        config,
	}
}

func NewEventBus(connection IDefaultKafkaConnection, retry int) cqrs.IEventBus {

	producer, err := sarama.NewAsyncProducerFromClient(connection.GetClient())
	if err != nil {
		logrus.Fatalln("creating Producer Error! %s", err.Error())
	}

	return BusKafka{connection, producer, cqrs.SubscriptionsManager, cqrs.EventBusConfig{
		retry,
		nil,
		nil,
		nil,
		cqrs.JsonMarshaler{},
	}}
}

func (c BusKafka) Subscribe(ctx context.Context, e event.IntegrationEvent, h event.IDynamicIntegrationEventHandler) {

	c.IEventBusSubscriptionsManager.AddSubscription(e)
	c.SubscribeDynamic(ctx, reflect.TypeOf(e).Name(), h)
}

func (c BusKafka) SubscribeDynamic(ctx context.Context, e string, h event.IDynamicIntegrationEventHandler) {

	c.IEventBusSubscriptionsManager.AddDynamicSubscription(e)
	cp, err := sarama.NewConsumerGroupFromClient(fmt.Sprintf("event_%s", e), c.connection.GetClient())
	if err != nil {
		return
	}

	go func() {
		if err := cp.Consume(ctx, []string{e}, consumer.NewDynamicIntegrationConsumerHandler(h, c.config)); err != nil {
			return
		}
	}()

	go func() {
		defer cp.Close()
	loop:
		for {
			select {
			case <-c.IEventBusSubscriptionsManager.GetHandle(e).UnSubscription:
				logrus.Infof("Unsubscribe events [%s]!\n", e)
				c.IEventBusSubscriptionsManager.RemoveDynamicSubscription(e)
				break loop
			}
		}
	}()
}

func (c BusKafka) SubscribeToDelay(e event.IntegrationEvent, h event.IDynamicIntegrationEventHandler) {
	//TODO implement me
	panic("implement me")
}

func (c BusKafka) UnSubscribe(e event.IntegrationEvent) {
	c.UnsubscribeDynamic(reflect.TypeOf(e).Name())
}

func (c BusKafka) UnsubscribeDynamic(e string) {
	c.IEventBusSubscriptionsManager.DynamicUnSubscription(e)

	//controller, err := c.connection.GetClient().Controller()
	//if err != nil {
	//	return
	//}
	//controller.DeleteGroups()
	//groups, err := controller.DescribeGroups(&sarama.DescribeGroupsRequest{
	//	Version:                     0,
	//	Groups:                      nil,
	//	IncludeAuthorizedOperations: false,
	//})
	//if err != nil {
	//	return
	//}
}

func (c BusKafka) Publish(e ...event.IntegrationEvent) {

	for _, i := range e {
		err := retry.Do(func() error {

			msg, err := c.config.Marshaler.Marshal(i)
			if err != nil {
				logrus.Errorf("Failed to serialize the Event Model: {%s}", msg.UUID)
				return err
			}

			bytes, err := json.Marshal(msg)
			if err != nil {
				logrus.Errorf("Failed to serialize the Message Model: {%s}", msg.UUID)
				return err
			}
			c.asyncProducer.Input() <- &sarama.ProducerMessage{Topic: reflect.TypeOf(i).Name(), Value: sarama.ByteEncoder(bytes)}
			logrus.Debugf("Publishing Event to Kafka: {%s}", msg.UUID)

		loop:
			for {
				select {
				case <-c.asyncProducer.Successes():
					break loop
				case err := <-c.asyncProducer.Errors():
					if err != nil {
						logrus.Errorf("Publishing Event to Kafka Error! {%s}", err.Err)
						return err
					}
				}
			}

			return nil

		}, retry.Attempts(uint(c.config.Retry)))

		if err != nil {
			logrus.Errorln(err.Error())
			return
		}
	}
}

func (c BusKafka) PublishToDelay(time time.Duration, e ...event.IntegrationEvent) {
	//TODO implement me
	panic("implement me")
}

func (c BusKafka) Disposable() error {

	err := c.asyncProducer.Close()
	c.connection.Close()
	c.Clear()
	return err
}
