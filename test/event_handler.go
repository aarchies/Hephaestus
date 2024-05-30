package test

import (
	"context"
	"github.com/sirupsen/logrus"
)

type TestHandler struct{}

func (t TestHandler) Handle(ctx context.Context, data interface{}) error {

	a := data.(TestModel)
	logrus.Infof("触发Handle %v\n", a)

	return nil
}
