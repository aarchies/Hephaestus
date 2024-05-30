package message

import (
	"time"
)

type Payload interface{}

type Message struct {
	// 消息Id
	UUID string
	// 元数据
	Metadata Metadata
	// 消息正文
	Payload Payload
	// 发送时间
	Time time.Time
}

func NewMessage(uuid string, metadata Metadata, payload []byte) *Message {

	return &Message{
		UUID:     uuid,
		Metadata: metadata,
		Payload:  payload,
		Time:     time.Now(),
	}
}
