package cqrs

import (
	"github.com/aarchies/hephaestus/messagec/cqrs/event"
	"reflect"
	"sync"
)

var SubscriptionsManager *IEventBusSubscriptionsManager

type IEventBusSubscriptionsManager struct {
	onEventRemoved OnEventRemove
	handlers       map[string]*Handler
	rwMutex        sync.RWMutex
}
type Handler struct {
	E              event.IntegrationEvent // event_interface
	UnSubscription chan int
	OnSubscription chan int
}
type OnEventRemove func(eventName string) error

func init() {
	SubscriptionsManager = &IEventBusSubscriptionsManager{
		handlers: make(map[string]*Handler),
	}
}

func (c *IEventBusSubscriptionsManager) GetHandle(name string) *Handler {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	if len(c.handlers) > 0 && c.handlers[name] != nil {
		return c.handlers[name]
	}

	return nil
}

func (c *IEventBusSubscriptionsManager) GetSubscription(eventName string) event.IntegrationEvent {
	c.rwMutex.RLock()
	defer c.rwMutex.RUnlock()

	if e := c.handlers[eventName].E; e != nil {
		return e
	}
	return nil
}

func (c *IEventBusSubscriptionsManager) AddSubscription(event event.IntegrationEvent) {
	c.AddDynamicSubscription(reflect.TypeOf(event).Name())
}

func (c *IEventBusSubscriptionsManager) AddDynamicSubscription(eventName string) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	if c.handlers[eventName] == nil {
		c.handlers[eventName] = &Handler{
			//E:              event,
			UnSubscription: make(chan int),
			OnSubscription: make(chan int),
		}
	}

	//c.handlers[eventName].E = event
	//c.handlers[eventName].OnSubscription <- 1
}

func (c *IEventBusSubscriptionsManager) DynamicUnSubscription(eventName string) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()

	c.handlers[eventName].UnSubscription <- 0
}

func (c *IEventBusSubscriptionsManager) UnSubscription(event event.IntegrationEvent) {
	c.DynamicUnSubscription(reflect.TypeOf(event).Name())
}

func (c *IEventBusSubscriptionsManager) RemoveDynamicSubscription(eventName string) {
	c.rwMutex.Lock()
	defer c.rwMutex.Unlock()
	delete(c.handlers, eventName)
}

func (c *IEventBusSubscriptionsManager) RemoveSubscription(event event.IntegrationEvent) {
	c.RemoveDynamicSubscription(reflect.TypeOf(event).Name())
}

func (c *IEventBusSubscriptionsManager) Clear() {
	clear(c.handlers)
}
