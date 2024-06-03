package event_bus

import (
	"github.com/aarchies/hephaestus/cqrs"
	"github.com/aarchies/hephaestus/cqrs/contrib/kafkax"
	"github.com/aarchies/hephaestus/examples/event_bus/pb"
	"github.com/aarchies/hephaestus/logs"
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
		Marshaler: cqrs.ProtobufMarshaler{}, // 指定序列化器
	})

	kafkax.Subscribe[TestModel, TestHandler]()
	time.Sleep(time.Second * 3)

	go func() {
		for i := 0; i < 10000; i++ {

			// protobuf
			message := TestModel{&pb.Weblog{
				Timestamp:       1717053897865,
				UserId:          "24",
				HostId:          "10334",
				NodeId:          "001",
				ReqId:           "fasggswraz",
				RemoteAddr:      "192.168.0.1",
				Protocol:        "",
				Method:          "",
				Host:            "",
				Uri:             "",
				Referer:         "",
				UserAgent:       "",
				ClientOs:        "",
				ClientFamily:    "",
				Crawler:         "",
				RequestSize:     0,
				CrwalerVerified: false,
				ResponseSize:    0,
				ResponseTime:    0,
				ResponseCode:    0,
				ResponseStatus:  "",
				UpstreamCode:    "",
				ContentType:     "",
				GeoContinent:    "",
				GeoCountry:      "",
				GeoRegion:       "",
				GeoCity:         "",
				GeoIsp:          "",
				GeoLat:          125.12,
				GeoLon:          122.12,
			}}
			// json
			//message := TestModel{"你好!"}
			kafkax.Publish(message)
			//time.Sleep(time.Second)
		}
	}()

	time.Sleep(time.Second * 10)
	kafkax.UnSubscribe[TestModel]()

	wg.Add(1)
	wg.Wait()
}
