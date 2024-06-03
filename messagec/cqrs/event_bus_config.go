package cqrs

import (
	"github.com/aarchies/hephaestus/messagec/cqrs/message"
	"github.com/google/uuid"
	"time"
)

type EventBusConfig struct {

	// 错误重试次数
	Retry int
	// 发布事件前触发
	OnPublishBefore OnEventSendFn
	// 发布事件后触发
	OnPublishAfter OnEventSendFn
	// 内部错误
	OnError OnEventErrorFn
	// 序列化器
	Marshaler Marshaler
}

type OnEventSendParams struct {
	UId       uuid.UUID
	EventName string
	Message   *message.Message
	time      time.Time
}
type OnEventSendFn func(params OnEventSendParams) error

type OnEventErrorParams struct {
	UId       uuid.UUID
	EventName string
	Message   *message.Message
	time      time.Time
	err       error
}
type OnEventErrorFn func(params OnEventErrorParams)
