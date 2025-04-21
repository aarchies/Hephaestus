package kafka

import (
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

func InitKafkaProducer(brokerServer []string, topic string, async bool) *kafka.Writer {
	logrus.Debugf("init kafka Producer brokerServer:%s,topic:[%s] v,async:[%v] \n", brokerServer, topic, async)

	w := &kafka.Writer{
		Addr:  kafka.TCP(brokerServer...),
		Topic: topic,
		// 分区选择策略（可选，默认 `RoundRobin`）
		// 可选值：
		// - kafka.RoundRobin：循环分配到不同分区
		// - kafka.Hash：基于 Key 的哈希值分配到特定分区
		// - kafka.LeastBytes：选择分区中最小负载的分区
		//Balancer:               &kafka.LeastBytes{},    // 指定分区的balancer模式为最小字节分布
		RequiredAcks:           kafka.RequireAll,       // ack模式
		Async:                  async,                  // 异步
		AllowAutoTopicCreation: true,                   // 是否允许创建不存在的topic
		MaxAttempts:            3,                      // 允许重试的次数（默认值为 10）
		BatchSize:              100,                    // 每次发送的最大批量消息数（默认值为 1）
		BatchTimeout:           100 * time.Millisecond, // 批量发送的时间间隔（默认值为 0，表示立即发送）
		BatchBytes:             10 * 1024 * 1024,       // 10MB 批量大小的最大字节数（默认值为 1MB）
		WriteTimeout:           10 * time.Second,       // 写入超时时间（默认值为 10 秒）
		// Compression: kafka.Snappy, // 是否启用消息压缩（默认不开启）可选值：None、Gzip、Snappy、Lz4、Zstd
	}

	return w
}
