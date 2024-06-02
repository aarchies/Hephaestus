package cqrs

import (
	"github.com/aarchies/hephaestus/messagec/cqrs/message"
)

type Marshaler interface {
	Marshal(v interface{}) ([]byte, string, error)
	Unmarshal(msg *message.Message, v interface{}) (err error)
}
