package cqrs

import (
	"github.com/aarchies/hephaestus/messagec/cqrs/event"
	"reflect"
	"slices"
	"sync"
)

var SubscriptionsManager *IEventBusSubscriptionsManager

type IEventBusSubscriptionsManager struct {
	onEventRemoved map[string]chan int
	handlers       map[string]*SubscriptionInfo
	eventType      []reflect.Type
	rwMutex        sync.RWMutex
}

func init() {
	SubscriptionsManager = &IEventBusSubscriptionsManager{
		handlers: make(map[string]*SubscriptionInfo),
	}
}

func (c *IEventBusSubscriptionsManager) IsEmpty() bool {
	return len(c.handlers) > 0
}
func (c *IEventBusSubscriptionsManager) Clear() {
	clear(c.handlers)
}

func (c *IEventBusSubscriptionsManager) AddDynamicSubscription(eventName string, handler event.IDynamicIntegrationEventHandler) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	if c.handlers[eventName] == nil {
		s := &SubscriptionInfo{}
		s.Dynamic(reflect.TypeOf(handler))

		c.handlers[eventName] = s
	}
}

func AddSubscription[T event.IntegrationEvent, TH event.IDynamicIntegrationEventHandler]() {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	eventName := reflect.TypeOf(event).Elem().Name()

	if c.handlers[eventName] == nil {
		s := &SubscriptionInfo{}
		s.Typed(reflect.TypeOf(handler))

		c.handlers[eventName] = s
	}

	if !slices.Contains(c.eventType, reflect.TypeOf(event)) {
		c.eventType = append(c.eventType, reflect.TypeOf(event))
	}

}

func (c *IEventBusSubscriptionsManager) GetHandler(eventName string) *SubscriptionInfo {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	if len(c.handlers) > 0 && c.handlers[eventName] != nil {
		return c.handlers[eventName]
	}

	return nil
}

func (c *IEventBusSubscriptionsManager) GetEventType(eventName string) reflect.Type {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	for _, r := range c.eventType {
		if r.Elem().Name() == eventName {
			return r
		}
	}

	return nil
}

func (c *IEventBusSubscriptionsManager) RemoveDynamicSubscription(eventName string) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	c.onEventRemoved[eventName] <- 0
	delete(c.handlers, eventName)
}

func (c *IEventBusSubscriptionsManager) RemoveSubscription(event event.IntegrationEvent) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	eventName := reflect.TypeOf(event).Elem().Name()
	c.onEventRemoved[eventName] <- 0
	delete(c.handlers, eventName)
	for index, i := range c.eventType {
		if i.Elem().Name() == eventName {
			c.eventType = append(c.eventType[:index], c.eventType[index+1:]...)
		}
	}
}
