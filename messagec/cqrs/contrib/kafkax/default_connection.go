package kafkax

import (
	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
	"time"
)

type IDefaultKafkaConnection interface {
	GetClient() sarama.Client
	GetConfig() *sarama.Config
	GetIsConnected() bool
	TryConnect() bool
	Close()
}

type DefaultKafkaConnection struct {
	c           Config
	config      *sarama.Config
	client      sarama.Client
	isConnected bool
}

func NewConnectionFactory(conf Config) IDefaultKafkaConnection {
	return &DefaultKafkaConnection{c: conf}
}

func (d *DefaultKafkaConnection) GetClient() sarama.Client {
	if !d.isConnected {
		d.TryConnect()
	}
	return d.client
}

func (d *DefaultKafkaConnection) GetConfig() *sarama.Config {
	return d.config
}

func (d *DefaultKafkaConnection) GetIsConnected() bool {
	return d.isConnected
}

func (d *DefaultKafkaConnection) TryConnect() bool {

	logrus.Infoln("Starting TopicModel KafKa.....")
	d.config = sarama.NewConfig()
	d.config.ClientID = d.c.ClientId
	d.config.Version = sarama.DefaultVersion
	d.config.Producer.Return.Successes = true
	d.config.Producer.Return.Errors = true
	d.config.Producer.RequiredAcks = sarama.WaitForAll
	d.config.Consumer.Offsets.Initial = sarama.OffsetOldest
	d.config.Consumer.Offsets.AutoCommit.Enable = false
	d.config.Consumer.Return.Errors = true
	d.config.Metadata.Full = false
	d.config.Net.SASL.User = d.c.Username
	d.config.Net.SASL.Password = d.c.Password
	d.config.Net.SASL.Enable = d.c.Username != "" && d.c.Password != ""
	d.config.Net.DialTimeout = 15 * time.Second

	c, err := sarama.NewClient(d.c.Adders, d.config)
	if err != nil {
		logrus.Fatalf("KafKa Client Connection Error! %s", err.Error())
		return false
	}
	d.client = c
	d.isConnected = !d.client.Closed()

	return d.isConnected
}

func (d *DefaultKafkaConnection) Close() {
	if err := d.client.Close(); err != nil {
		logrus.Errorf(err.Error())
		return
	}
}
