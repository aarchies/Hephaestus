package cqrs

import (
	"github.com/aarchies/hephaestus/cqrs/event"
	"github.com/sirupsen/logrus"
	"reflect"
	"slices"
)

var SubscriptionsManager *IEventBusSubscriptionsManager

type IEventBusSubscriptionsManager struct {
	OnEventRemoved map[string]chan int
	handlers       map[string]*SubscriptionInfo
	eventType      []reflect.Type
	//rwMutex        sync.RWMutex
}

func init() {
	SubscriptionsManager = &IEventBusSubscriptionsManager{
		handlers:       make(map[string]*SubscriptionInfo),
		OnEventRemoved: make(map[string]chan int),
	}
}

func (c *IEventBusSubscriptionsManager) IsEmpty() bool {
	return len(c.handlers) > 0
}

func (c *IEventBusSubscriptionsManager) Clear() {
	clear(c.handlers)
	clear(c.eventType)
}

func (c *IEventBusSubscriptionsManager) GetHandler(eventName string) *SubscriptionInfo {

	if len(c.handlers) > 0 && c.handlers[eventName] != nil {
		return c.handlers[eventName]
	}

	return nil
}

func (c *IEventBusSubscriptionsManager) GetEventType(eventName string) reflect.Type {

	for _, r := range c.eventType {
		if r.Elem().Name() == eventName {
			return r
		}
	}

	return nil
}

func AddDynamicSubscription[TH event.IDynamicIntegrationEventHandler](eventName string) {

	if SubscriptionsManager.handlers[eventName] == nil {
		SubscriptionsManager.OnEventRemoved[eventName] = make(chan int)
		SubscriptionsManager.handlers[eventName] = &SubscriptionInfo{
			IsDynamic:   true,
			HandlerType: reflect.TypeOf(new(TH)),
		}
	} else {
		logrus.Fatalf("Current events have been subscribed! Dynamic Handler:{%s}\n", SubscriptionsManager.handlers[eventName].HandlerType.Elem().Name())
	}
}

func AddSubscription[T event.IntegrationEvent, TH event.IntegrationEventHandler[T]]() {

	eventName := reflect.TypeOf(new(T)).Elem().Name()

	if SubscriptionsManager.handlers[eventName] == nil {
		SubscriptionsManager.OnEventRemoved[eventName] = make(chan int)
		SubscriptionsManager.handlers[eventName] = &SubscriptionInfo{
			IsDynamic:   false,
			HandlerType: reflect.TypeOf(new(TH)),
		}
	} else {
		logrus.Fatalln("Current events have been subscribed! Handler:{%s}\n", SubscriptionsManager.handlers[eventName].HandlerType.Elem().Name())
	}

	if !slices.Contains(SubscriptionsManager.eventType, reflect.TypeOf(new(T))) {
		SubscriptionsManager.eventType = append(SubscriptionsManager.eventType, reflect.TypeOf(new(T)))
	}

}

func RemoveDynamicSubscription(eventName string) {

	SubscriptionsManager.OnEventRemoved[eventName] <- -1
	//delete(SubscriptionsManager.handlers, eventName)
}

func RemoveSubscription[T event.IntegrationEvent]() {

	eventName := reflect.TypeOf(new(T)).Elem().Name()
	SubscriptionsManager.OnEventRemoved[eventName] <- -1
	//delete(SubscriptionsManager.handlers, eventName)
	//for index, i := range SubscriptionsManager.eventType {
	//	if i.Elem().Name() == eventName {
	//		SubscriptionsManager.eventType = append(SubscriptionsManager.eventType[:index], SubscriptionsManager.eventType[index+1:]...)
	//	}
	//}
}
