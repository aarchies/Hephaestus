package event_bus

import (
	"github.com/aarchies/hephaestus/logs"
	"github.com/aarchies/hephaestus/messagec/cqrs"
	"github.com/aarchies/hephaestus/messagec/cqrs/contrib/kafkax"
	"github.com/sirupsen/logrus"
	"sync"
	"testing"
	"time"
)

func Test(t *testing.T) {
	logs.SetLogsModule(logrus.DebugLevel)

	var wg sync.WaitGroup

	factory := kafkax.NewConnectionFactory(kafkax.NewConfig([]string{"154.88.24.86:19092", "154.88.24.86:29092", "154.88.24.86:39092"}, "clickId", "", ""))
	kafkax.NewEventBusWithConfig(factory, cqrs.EventBusConfig{
		Retry:           3,
		OnPublishBefore: nil,
		OnPublishAfter:  nil,
		OnError: func(params cqrs.OnEventErrorParams) {
			logrus.Errorln(params)
		},
		Marshaler: cqrs.JsonMarshaler{},
	})

	kafkax.Subscribe[TestModel, TestHandler]()

	go func() {
		for i := 0; i < 100; i++ {
			kafkax.Publish(TestModel{Data: "hello"})
		}
	}()

	time.Sleep(time.Second * 7)
	//caller.UnSubscribe(NewTestModel())

	wg.Add(1)
	wg.Wait()
}
