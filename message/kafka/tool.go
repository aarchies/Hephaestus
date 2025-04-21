package kafka

import (
	"context"
	"fmt"
	"log"

	sarama "github.com/IBM/sarama"
	"github.com/vmihailenco/msgpack/v5"
)

// KafkaManager holds the Kafka producer and consumer instances.
type KafkaManager struct {
	Producer sarama.SyncProducer
	Consumer sarama.Consumer
}

// NewKafkaManager creates a new KafkaManager instance.
func NewKafkaManager(brokerList []string, topic string) (*KafkaManager, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	// Create producer
	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %v", err)
	}

	// Create consumer
	consumer, err := sarama.NewConsumer(brokerList, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka consumer: %v", err)
	}

	return &KafkaManager{
		Producer: producer,
		Consumer: consumer,
	}, nil
}

// Close closes the Kafka producer and consumer.
func (km *KafkaManager) Close() {
	if err := km.Producer.Close(); err != nil {
		log.Printf("Error closing Kafka producer: %v\n", err)
	}
	if err := km.Consumer.Close(); err != nil {
		log.Printf("Error closing Kafka consumer: %v\n", err)
	}
}

// Produce sends a message to Kafka.
func (km *KafkaManager) Produce(ctx context.Context, key string, value interface{}, topic string) error {

	// Serialize message using msgpack
	msgBytes, err := msgpack.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to serialize message: %v", err)
	}

	// Create producer message
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(msgBytes),
	}

	// Send message
	go km.Producer.SendMessage(msg)

	//fmt.Printf("Produced message to topic %s, partition %d, offset %d\n", topic, partition, offset)
	//
	//elapsed := time.Since(start)
	//fmt.Printf("Program took %s\n", elapsed)

	return nil
}

// Consume reads messages from Kafka.
func (km *KafkaManager) Consume(ctx context.Context, topic string) error {
	// Create partition consumer for topic
	partitionConsumer, err := km.Consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		return fmt.Errorf("failed to start consumer for topic %s: %v", topic, err)
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Printf("Error closing partition consumer: %v\n", err)
		}
	}()

	// Handle messages
	for {
		select {
		case msg := <-partitionConsumer.Messages():
			var value interface{} // Change to your struct type
			if err := msgpack.Unmarshal(msg.Value, &value); err != nil {
				return fmt.Errorf("failed to deserialize message: %v", err)
			}
			fmt.Printf("Message received: %+v\n", value)

			// Commit message offset
			//partitionConsumer.MarkOffset(msg, "")
		case err := <-partitionConsumer.Errors():
			return fmt.Errorf("consumer error: %v", err)
		case <-ctx.Done():
			return nil
		}
	}
}

// Example usage:
//
// km, err := NewKafkaManager([]string{"localhost:9092"}, "test-topic")
// if err != nil {
//     log.Fatalf("Failed to create Kafka manager: %v\n", err)
// }
// defer km.Close()
//
// // Produce example:
// err := km.Produce(context.Background(), "key-A", Person{Name: "John Doe", Age: 30, Address: "123 Elm Street"})
// if err != nil {
//     log.Fatalf("Failed to produce message: %v", err)
// }
//
// // Consume example:
// go func() {
//     err := km.Consume(context.Background(), "test-topic")
//     if err != nil {
//         log.Fatalf("Failed to consume messages: %v", err)
//     }
// }()

func TestConsume(brokerList []string, topic string) {
	km, err := NewKafkaManager(brokerList, topic)
	if err != nil {
		log.Fatalf("Failed to create Kafka manager: %v\n", err)
	}
	//defer km.Close()

	// Consume example:
	go func() {
		err := km.Consume(context.Background(), topic)
		if err != nil {
			log.Fatalf("Failed to consume messages: %v", err)
		}
	}()
}
