package cqrs

import (
	"github.com/aarchies/hephaestus/cqrs/message"
	"reflect"
)

type Marshaler interface {
	Marshal(v interface{}) ([]byte, string, error)
	Unmarshal(msg *message.Message, v reflect.Value) (err error)
}
