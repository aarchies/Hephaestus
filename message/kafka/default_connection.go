package kafka

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/sirupsen/logrus"
)

type IDefaultKafkaConnection interface {
	GetClient() sarama.Client
	GetConfig() *sarama.Config
	GetProducer() sarama.AsyncProducer
	GetIsConnected() bool
	TryConnect() bool
	Close()
}

var DefaultClient IDefaultKafkaConnection

type DefaultKafkaConnection struct {
	c           Config
	config      *sarama.Config
	client      sarama.Client
	producer    sarama.AsyncProducer
	isConnected bool
}

func (d *DefaultKafkaConnection) GetProducer() sarama.AsyncProducer {
	return d.producer
}

func NewConnectionFactory(conf Config) IDefaultKafkaConnection {
	d := &DefaultKafkaConnection{c: conf}
	d.TryConnect()

	p, err := sarama.NewAsyncProducerFromClient(d.GetClient())
	if err != nil {
		panic(err)
	}
	d.producer = p

	DefaultClient = d
	return d
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

	logrus.Infof("start kafka cluster %s username:[%s] password:[%s] connecting...\n", d.c.Adders, d.c.Username, d.c.Password)
	d.config = sarama.NewConfig()
	d.config.ClientID = d.c.ClientId
	d.config.Version = sarama.DefaultVersion
	d.config.Producer.Return.Successes = true
	d.config.Producer.Return.Errors = true
	d.config.Producer.RequiredAcks = sarama.WaitForAll

	d.config.Consumer.Offsets.AutoCommit.Enable = true
	d.config.Consumer.Offsets.AutoCommit.Interval = time.Second
	d.config.Consumer.Offsets.Initial = sarama.OffsetOldest

	d.config.Consumer.Offsets.Retry.Max = 3
	d.config.Consumer.Return.Errors = true
	d.config.Metadata.Full = false
	d.config.Net.SASL.User = d.c.Username
	d.config.Net.SASL.Password = d.c.Password
	d.config.Net.SASL.Enable = d.c.Username != "" && d.c.Password != ""
	//d.config.Net.DialTimeout = 3 * time.Second

	c, err := sarama.NewClient(d.c.Adders, d.config)
	if err != nil {
		logrus.Fatalf("kafKa client connection error! %s", err.Error())
		return false
	}
	d.client = c
	d.isConnected = !d.client.Closed()

	if err := c.RefreshMetadata(); err != nil {
		logrus.Fatalf("error connecting to kafka cluster: %v", err)
	}

	logrus.Infoln("kafka client connected successful!")

	return d.isConnected
}

func (d *DefaultKafkaConnection) Close() {
	if err := d.client.Close(); err != nil {
		logrus.Errorf(err.Error())
		return
	}
	d.producer.Close()
}
