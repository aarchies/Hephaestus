package test

import (
	"context"
	"github.com/sirupsen/logrus"
)

type EventHandler struct{}

func (t EventHandler) Handle(ctx context.Context, data interface{}) error {

	a := data.(EventModel)
	logrus.Infof("触发Handle %v\n", a)

	return nil
}
