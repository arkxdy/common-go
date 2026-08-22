package kafka

import (
	"log"
	"time"

	"github.com/IBM/sarama"
)

// Config holds common Kafka settings.
type Config struct {
	Brokers []string
	Version sarama.KafkaVersion
}

// ProducerConfig holds settings for the synchronous producer.
type ProducerConfig struct {
	Config
	Idempotent      bool
	RequiredAcks    sarama.RequiredAcks
	RetryMax        int
	RetryBackoff    time.Duration
	ReturnSuccesses bool
	ReturnErrors    bool
	MaxOpenRequests int
}

// DefaultProducerConfig returns a ProducerConfig with safe defaults.
func DefaultProducerConfig(brokers []string) ProducerConfig {
	return ProducerConfig{
		Config: Config{
			Brokers: brokers,
			Version: sarama.DefaultVersion,
		},
		Idempotent:      true,
		RequiredAcks:    sarama.WaitForAll,
		RetryMax:        5,
		RetryBackoff:    100 * time.Millisecond,
		ReturnSuccesses: true,
		ReturnErrors:    true,
		MaxOpenRequests: 1,
	}
}

// ConsumerConfig holds settings for the consumer group.
type ConsumerConfig struct {
	Config
	GroupID       string
	Topics        []string
	InitialOffset int64
	Handler       MessageHandler
	ErrorLogger   *log.Logger
}

// DefaultConsumerConfig returns a ConsumerConfig with safe defaults.
func DefaultConsumerConfig(brokers []string, groupID string, topics []string, handler MessageHandler) ConsumerConfig {
	return ConsumerConfig{
		Config: Config{
			Brokers: brokers,
			Version: sarama.DefaultVersion,
		},
		GroupID:       groupID,
		Topics:        topics,
		InitialOffset: sarama.OffsetOldest,
		Handler:       handler,
		ErrorLogger:   log.Default(),
	}
}
