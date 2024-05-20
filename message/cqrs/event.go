package cqrs

import "github.com/google/uuid"

type Event interface {
	GetID() uuid.UUID
}
