package cqrs

import "flow_crafter_CDN/pkg/messagec/cqrs/message"

type Marshaler interface {
	Marshal(v interface{}) (*message.Message, error)
	Unmarshal(msg *message.Message, v interface{}) (err error)
}
