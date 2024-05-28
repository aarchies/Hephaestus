package event_bus

import (
	"context"
	"github.com/aarchies/go-lib/examples/event_bus/event"
	"github.com/aarchies/go-lib/logs"
	"github.com/aarchies/go-lib/messagec/cqrs/contrib/kafka"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

func main() {
	logs.AddLogsModule(logrus.DebugLevel)
	var wg sync.WaitGroup
	factory := kafka.NewConnectionFactory(kafka.NewConfig([]string{"ip:port", "ip:port", "ip:port"}, "clickId", "", ""))
	c := kafka.NewEventBus(factory, 3)
	//c := kafka.NewEventBusWithConfig(factory, cqrs.EventBusConfig{
	//	Retry:           3,
	//	OnPublishBefore: nil,
	//	OnPublishAfter:  nil,
	//	OnError: func(params cqrs.OnEventErrorParams) {
	//		logrus.Errorln(params)
	//	},
	//	Marshaler: cqrs.JsonMarshaler{},
	//})

	// 订阅
	e := event.NewTestModel()

	c.Subscribe(context.Background(), e, event.TestHandler{})

	go func() {
		for i := 0; i < 200; i++ {
			c.Publish(event.NewTestModel())
			time.Sleep(time.Second * 1)
		}
	}()

	time.Sleep(time.Second * 7)

	// 取消订阅
	c.UnSubscribe(event.NewTestModel())

	wg.Add(3)
	wg.Wait()
}
