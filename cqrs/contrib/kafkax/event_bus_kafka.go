package kafkax

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/aarchies/hephaestus/cqrs"
	"github.com/aarchies/hephaestus/cqrs/contrib/kafkax/consumer"
	"github.com/aarchies/hephaestus/cqrs/event"
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

	eventName := reflect.TypeOf(new(T)).Elem().Name()
	logrus.Infof("Subscribing to event {%s} with {%s}!", eventName, reflect.TypeOf(new(TH)).Elem().Name())
	cqrs.AddSubscription[T, TH]()

	cp, err := sarama.NewConsumerGroupFromClient(fmt.Sprintf("event_%s", eventName), c.connection.GetClient())
	if err != nil {
		return
	}

	go func() {
		if err := cp.Consume(context.Background(), []string{eventName}, consumer.NewIntegrationConsumerHandler(c.subscriptionsManager, c.config)); err != nil {
			return
		}
	}()

	go func() {
		select {
		case <-c.subscriptionsManager.OnEventRemoved[eventName]:
			logrus.Infof("Unsubscribe to events [%s]!\n", eventName)
			cp.Close()
			return
		}
	}()
}

func SubscribeDynamic[TH event.IDynamicIntegrationEventHandler](e string) {

	logrus.Infof("Subscribing Dynamic to event {%s} with {%s}", e, reflect.TypeOf(new(TH)).Elem().Name())
	cqrs.AddDynamicSubscription[TH](e)
	cp, err := sarama.NewConsumerGroupFromClient(fmt.Sprintf("event_%s", e), c.connection.GetClient())
	if err != nil {
		return
	}

	go func() {

		if err := cp.Consume(context.Background(), []string{e}, consumer.NewDynamicIntegrationConsumerHandler(c.subscriptionsManager, c.config)); err != nil {
			return
		}
	}()

	go func() {
		for {
			select {
			case <-c.subscriptionsManager.OnEventRemoved[e]:
				logrus.Infof("Unsubscribe Dynamic events [%s]!\n", e)
				cp.Close()
				return
			}
		}
	}()
}

func UnSubscribe[T event.IntegrationEvent]() {
	cqrs.RemoveSubscription[T]()
}

func UnSubscribeDynamic(e string) {
	cqrs.RemoveDynamicSubscription(e)
	logrus.Infof("Unsubscribed Dynamic to event {%s}", e)
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

		loop:
			for {
				select {
				case <-c.asyncProducer.Successes():
					logrus.Debugf("Publishing Event to Kafka: {%s}", uid)
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

func PublishToDelay(times time.Duration, e ...event.IntegrationEvent) {
	for _, i := range e {
		time.Sleep(times)
		err := retry.Do(func() error {

			bytes, uid, err := c.config.Marshaler.Marshal(i)
			if err != nil {
				logrus.Errorf("Failed to serialize the Event Model: {%s}", i.GetId())
				return err
			}

			c.asyncProducer.Input() <- &sarama.ProducerMessage{Topic: reflect.TypeOf(i).Name(), Value: sarama.ByteEncoder(bytes)}

		loop:
			for {
				select {
				case <-c.asyncProducer.Successes():
					logrus.Debugf("Publishing Event to Kafka: {%s}", uid)
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

func Disposable() error {
	err := c.asyncProducer.Close()
	c.connection.Close()
	c.subscriptionsManager.Clear()
	return err
}
