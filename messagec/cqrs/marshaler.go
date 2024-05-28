package cqrs

import "github.com/aarchies/go-lib/messagec/cqrs/message"

type Marshaler interface {
	Marshal(v interface{}) (*message.Message, error)
	Unmarshal(msg *message.Message, v interface{}) (err error)
}
