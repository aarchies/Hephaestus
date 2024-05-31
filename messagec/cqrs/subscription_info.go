package cqrs

import "reflect"

type SubscriptionInfo struct {
	IsDynamic   bool
	HandlerType reflect.Type
}

func (c SubscriptionInfo) Dynamic(h reflect.Type) {
	c.IsDynamic = true
	c.HandlerType = h

}

func (c SubscriptionInfo) Typed(h reflect.Type) {
	c.IsDynamic = false
	c.HandlerType = h

}
