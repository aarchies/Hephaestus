package event_bus

import (
	"context"
	"github.com/aarchies/hephaestus/examples/event_bus/event"
	"github.com/aarchies/hephaestus/logs"
	"github.com/aarchies/hephaestus/messagec/cqrs"
	"github.com/aarchies/hephaestus/messagec/cqrs/contrib/kafka"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

func main() {
	logs.AddLogsModule(logrus.DebugLevel)

	var wg sync.WaitGroup
	factory := kafka.NewConnectionFactory(kafka.NewConfig([]string{"ip:port", "ip:port", "ip:port"}, "clickId", "", ""))
	//c := kafka.NewEventBus(factory, 3)
	c := kafka.NewEventBusWithConfig(factory, cqrs.EventBusConfig{
		Retry:           3,
		OnPublishBefore: nil,
		OnPublishAfter:  nil,
		OnError: func(params cqrs.OnEventErrorParams) {
			logrus.Errorln(params)
		},
		Marshaler: cqrs.JsonMarshaler{},
	})

	c.Subscribe(context.Background(), event.NewTestModel(), event.TestHandler{})

	go func() {
		for i := 0; i < 200; i++ {
			c.Publish(event.NewTestModel())
			time.Sleep(time.Second * 1)
		}
	}()

	time.Sleep(time.Second * 7)
	c.UnSubscribe(event.NewTestModel())

	wg.Add(1)
	wg.Wait()
}
