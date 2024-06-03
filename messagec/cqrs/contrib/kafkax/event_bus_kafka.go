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

func NewEventBus(connection IDefaultKafkaConnection) BusKafka {

	producer, err := sarama.NewAsyncProducerFromClient(connection.GetClient())
	if err != nil {
		logrus.Fatalln("creating Producer Error! %s", err.Error())
	}
	c = BusKafka{connection: connection, asyncProducer: producer, subscriptionsManager: cqrs.SubscriptionsManager, config: cqrs.EventBusConfig{
		3,
		nil,
		nil,
		nil,
		nil,
	}}
	return c
}
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

func Subscribe[T event.IntegrationEvent, TH event.IntegrationEventHandler[T]]() {

	cqrs.AddSubscription[T, TH]()

	eventName := reflect.TypeOf(new(T)).Elem().Name()
	handlerName := reflect.TypeOf(new(TH)).Elem().Name()
	cp, err := sarama.NewConsumerGroupFromClient(fmt.Sprintf("event_%s", eventName), c.connection.GetClient())
	if err != nil {
		return
	}

	logrus.Infof("Subscribing to event {%s} with {%s}", eventName, handlerName)
	go func() {
		if err := cp.Consume(context.Background(), []string{eventName}, consumer.NewIntegrationConsumerHandler(c.subscriptionsManager, c.config)); err != nil {
			return
		}
	}()

	//go func() {
	//	defer cp.Close()
	//loop:
	//	for {
	//		select {
	//		case <-c.subscriptionsManager.OnEventRemoved[eventName]:
	//			logrus.Infof("Unsubscribe events [%s]!\n", eventName)
	//			c.subscriptionsManager.RemoveDynamicSubscription(eventName)
	//			break loop
	//		}
	//	}
	//}()
}
func SubscribeDynamic[TH event.IDynamicIntegrationEventHandler](e string) {
	cqrs.AddDynamicSubscription[TH](e)

	handlerName := reflect.TypeOf(new(TH)).Elem().Name()
	cp, err := sarama.NewConsumerGroupFromClient(fmt.Sprintf("event_%s", e), c.connection.GetClient())
	if err != nil {
		return
	}

	logrus.Infof("Subscribing to event {%s} with {%s}", e, handlerName)
	go func() {
		if err := cp.Consume(context.Background(), []string{e}, consumer.NewDynamicIntegrationConsumerHandler(c.subscriptionsManager, c.config)); err != nil {
			return
		}
	}()

	//go func() {
	//	defer cp.Close()
	//loop:
	//	for {
	//		select {
	//		case <-c.subscriptionsManager.OnEventRemoved[eventName]:
	//			logrus.Infof("Unsubscribe events [%s]!\n", eventName)
	//			c.subscriptionsManager.RemoveDynamicSubscription(eventName)
	//			break loop
	//		}
	//	}
	//}()
}
func SubscribeToDelay[T event.IntegrationEvent, TH event.IntegrationEventHandler[T]]() {
	//TODO implement me
	panic("implement me")
}

func (c BusKafka) UnSubscribe(e event.IntegrationEvent) {
	//c.subscriptionsManager.DynamicUnSubscription(e)
	//c.UnsubscribeDynamic(reflect.TypeOf(e).Name())
	logrus.Infof("Unsubscribed to event {%s}", e)
}

func Publish(e ...event.IntegrationEvent) {

	for _, i := range e {
		err := retry.Do(func() error {

			bytes, uid, err := c.config.Marshaler.Marshal(i)
			if err != nil {
				logrus.Errorf("Failed to serialize the Event Model: {%s}", i.GetId())
				return err
			}

			c.asyncProducer.Input() <- &sarama.ProducerMessage{Topic: reflect.TypeOf(i).Name(), Value: sarama.ByteEncoder(bytes)}
			logrus.Debugf("Publishing Event to Kafka: {%s}", uid)

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
func PublishToDelay[T event.IntegrationEvent](time time.Duration, e ...T) {
	//TODO implement me
	return
}

func Disposable() error {
	err := c.asyncProducer.Close()
	c.connection.Close()
	return err
}
