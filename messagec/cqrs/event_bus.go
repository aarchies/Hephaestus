package cqrs

// IEventBus 事件基本模型
type IEventBus interface {
	//// Subscribe 事件模型订阅
	//Subscribe(ctx context.Context, e event.IntegrationEvent, h event.IntegrationEventHandler)
	//// SubscribeDynamic 根据事件名称进行订阅
	//SubscribeDynamic(ctx context.Context, e string, h event.IntegrationEventHandler)
	//// SubscribeToDelay 延时订阅
	//SubscribeToDelay(e event.IntegrationEvent, h event.IntegrationEventHandler)
	//// UnSubscribe 取消订阅
	//UnSubscribe(e event.IntegrationEvent)
	//// UnsubscribeDynamic 根据事件名称取消订阅
	//UnsubscribeDynamic(e string)
	//// Publish 推送事件
	//Publish(e ...event.IntegrationEvent)
	//// PublishToDelay 延时发布
	//PublishToDelay(time time.Duration, e ...event.IntegrationEvent)
	//// Disposable 销毁资源
	//Disposable() error
}
