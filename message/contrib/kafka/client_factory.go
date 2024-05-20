package kafka

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
	"time"
)

var conf *sarama.Config
var hosts []string

type Cli struct {
	Client sarama.Client
}

func Register(adders []string, clientId string, username string, password string) sarama.Client {
	fmt.Printf("load kafka config host:%v\n", adders)

	config := sarama.NewConfig()
	config.ClientID = clientId
	config.Version = sarama.DefaultVersion           // kafka server的版本号
	config.Producer.Return.Successes = true          // sync必须设置这个
	config.Producer.RequiredAcks = sarama.WaitForAll // 也就是等待follower同步，才会返回
	config.Producer.Retry = struct {
		Max         int
		Backoff     time.Duration
		BackoffFunc func(retries int, maxRetries int) time.Duration
	}{
		Max:         3,
		Backoff:     500 * time.Millisecond,
		BackoffFunc: nil,
	}
	config.Producer.Return.Errors = true
	config.Consumer.Return.Errors = true
	config.Metadata.Full = true
	config.Consumer.Offsets.AutoCommit.Enable = true          // 自动提交偏移量，默认开启
	config.Consumer.Offsets.AutoCommit.Interval = time.Second // commit提交频率，不然容易down机后造成重复消费。
	config.Consumer.Offsets.Initial = sarama.OffsetOldest     // 从最开始的地方消费，业务中看有没有需求，新业务重跑topic。
	config.Net.SASL.User = username
	config.Net.SASL.Password = password
	config.Net.SASL.Enable = username != "" && password != ""

	conf = config
	hosts = adders

	c, err := sarama.NewClient(hosts, conf)
	if err != nil {
		logrus.Fatalln("连接kafka失败!" + err.Error())
	}
	return c
}

func GetClient() Cli {

	c, err := sarama.NewClient(hosts, conf)
	if err != nil {
		logrus.Fatalln("连接kafka失败!" + err.Error())
	}
	return Cli{
		Client: c,
	}
}

func (c *Cli) Producer(topic string, data interface{}, isDebug bool) {

	producer, err := sarama.NewAsyncProducerFromClient(c.Client)
	if err != nil {
		return
	}

	defer func() {
		if err := producer.Close(); err != nil {
			logrus.Error("Failed to shut down producer cleanly:", err)
		}
	}()

	msg := &sarama.ProducerMessage{Topic: topic, Value: sarama.ByteEncoder(data.(string))}
	producer.Input() <- msg

loop:
	for {
		select {
		case <-producer.Successes():
			if isDebug {
				logrus.Debugf("Successfully send message to topic %s: %s", topic, data.(string))
			}
			break loop
		case err := <-producer.Errors():
			if err != nil {
				logrus.Errorf("Failed to send message to topic %s: %v", topic, err.Err)
				break loop
			}
		}
	}
}
