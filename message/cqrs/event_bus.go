package cqrs

import (
	"context"
	"time"
)

type IEventBus interface {
	Subscribe()
	SubscribeDynamic(topic ...string)
	UnSubscribe()
	UnsubscribeDynamic(topic ...string)
	Publish(ctx context.Context, msg ...Event)
	PublishToDelay(ctx context.Context, time time.Duration, msg ...Event)
	HealthyCheck()
}

type genericCommandHandler[Command any] struct {
	handleFunc  func(ctx context.Context, cmd *Command) error
	handlerName string
}

func (g genericCommandHandler[Command]) Subscribe() {
	//TODO implement me
	panic("implement me")
}

func (g genericCommandHandler[Command]) SubscribeDynamic(topic ...string) {
	//TODO implement me
	panic("implement me")
}

func (g genericCommandHandler[Command]) UnSubscribe() {
	//TODO implement me
	panic("implement me")
}

func (g genericCommandHandler[Command]) UnsubscribeDynamic(topic ...string) {
	//TODO implement me
	panic("implement me")
}

func (g genericCommandHandler[Command]) Publish(ctx context.Context, msg ...Event) {
	//TODO implement me
	panic("implement me")
}

func (g genericCommandHandler[Command]) PublishToDelay(ctx context.Context, time time.Duration, msg ...Event) {
	//TODO implement me
	panic("implement me")
}

func (g genericCommandHandler[Command]) HealthyCheck() {
	//TODO implement me
	panic("implement me")
}
