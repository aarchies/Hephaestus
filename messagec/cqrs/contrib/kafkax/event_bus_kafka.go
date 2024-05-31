package kafkax

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/aarchies/hephaestus/messagec/cqrs"
	"github.com/aarchies/hephaestus/messagec/cqrs/contrib/kafkax/consumer"

	"github.com/aarchies/hephaestus/messagec/cqrs/event"
	"github.com/avast/retry-go"
	"github.com/sirupsen/logrus"
	"reflect"
	"time"
)

type BusKafka struct {
	connection           IDefaultKafkaConnection
	asyncProducer        sarama.AsyncProducer
	subscriptionsManager *cqrs.IEventBusSubscriptionsManager
	config               cqrs.EventBusConfig
}

var c BusKafka

func NewEventBusWithConfig(connection IDefaultKafkaConnection, config cqrs.EventBusConfig) BusKafka {
	producer, err := sarama.NewAsyncProducerFromClient(connection.GetClient())
	if err != nil {
		logrus.Fatalln("creating Producer Error! %s", err.Error())
	}

	c = BusKafka{
		connection:           connection,
		asyncProducer:        producer,
		subscriptionsManager: cqrs.SubscriptionsManager,
		config:               config,
	}
	return c
}

func NewEventBus(connection IDefaultKafkaConnection, retry int) BusKafka {

	producer, err := sarama.NewAsyncProducerFromClient(connection.GetClient())
	if err != nil {
		logrus.Fatalln("creating Producer Error! %s", err.Error())
	}
	c = BusKafka{connection: connection, asyncProducer: producer, subscriptionsManager: cqrs.SubscriptionsManager, config: cqrs.EventBusConfig{
		retry,
		nil,
		nil,
		nil,
		nil,
	}}
	return c
}

func Subscribe[T event.IntegrationEvent, TH event.IDynamicIntegrationEventHandler]() {

	e := new(T)
	h := new(TH)
	c.subscriptionsManager.AddSubscription(e, h)
	SubscribeDynamic[TH](reflect.TypeOf(new(T)).Elem().Name())

	c.subscriptionsManager.AddDynamicSubscription(e)
	cp, err := sarama.NewConsumerGroupFromClient(fmt.Sprintf("event_%s", e), c.connection.GetClient())
	if err != nil {
		return
	}

	logrus.Infof("Subscribing to event {%s} with {%s}", e, reflect.TypeOf(new(TH)).Elem().Name())
	go func() {
		if err := cp.Consume(context.Background(), []string{e}, consumer.DynamicIntegrationConsumerHandler(h, c.config)); err != nil {
			return
		}
	}()

	go func() {
		defer cp.Close()
	loop:
		for {
			select {
			case <-c.subscriptionsManager.GetHandle(e).UnSubscription:
				logrus.Infof("Unsubscribe events [%s]!\n", e)
				c.subscriptionsManager.RemoveDynamicSubscription(e)
				break loop
			}
		}
	}()
}

func SubscribeDynamic[TH event.IDynamicIntegrationEventHandler](e string) {

	c.subscriptionsManager.AddDynamicSubscription(e)
	cp, err := sarama.NewConsumerGroupFromClient(fmt.Sprintf("event_%s", e), c.connection.GetClient())
	if err != nil {
		return
	}

	logrus.Infof("Subscribing to event {%s} with {%s}", e, reflect.TypeOf(new(TH)).Elem().Name())
	go func() {
		if err := cp.Consume(context.Background(), []string{e}, consumer.DynamicIntegrationConsumerHandler(h, c.config)); err != nil {
			return
		}
	}()

	go func() {
		defer cp.Close()
	loop:
		for {
			select {
			case <-c.subscriptionsManager.GetHandle(e).UnSubscription:
				logrus.Infof("Unsubscribe events [%s]!\n", e)
				c.subscriptionsManager.RemoveDynamicSubscription(e)
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
	c.subscriptionsManager.DynamicUnSubscription(e)
	logrus.Infof("Unsubscribed to event {%s}", e)
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

			bytes, err := c.config.Marshaler.Marshal(i)
			if err != nil {
				logrus.Errorf("Failed to serialize the Event Model: {%s}", i.GetId())
				return err
			}

			c.asyncProducer.Input() <- &sarama.ProducerMessage{Topic: reflect.TypeOf(i).Elem().Name(), Value: sarama.ByteEncoder(bytes)}
			logrus.Debugf("Publishing Event to Kafka: {%s}", i.GetId())

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
	return err
}
