package kafka

import (
	"github.com/IBM/sarama"
)

// Producer wraps a sarama.SyncProducer and provides a simple, configurable API.
type Producer struct {
	producer sarama.SyncProducer
}

// NewProducer creates a new synchronous Kafka producer from the given config.
func NewProducer(cfg ProducerConfig) (*Producer, error) {
	config := sarama.NewConfig()
	config.Version = cfg.Version

	config.Producer.Idempotent = cfg.Idempotent
	config.Producer.RequiredAcks = cfg.RequiredAcks
	config.Producer.Retry.Max = cfg.RetryMax
	config.Producer.Retry.Backoff = cfg.RetryBackoff
	config.Producer.Return.Successes = cfg.ReturnSuccesses
	config.Producer.Return.Errors = cfg.ReturnErrors
	config.Net.MaxOpenRequests = cfg.MaxOpenRequests

	p, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{producer: p}, nil
}

// SendMessage publishes a message to Kafka and returns the partition and offset.
// Headers are optional and can be provided as variadic arguments.
func (p *Producer) SendMessage(topic string, key, value []byte, headers ...sarama.RecordHeader) (partition int32, offset int64, err error) {
	msg := &sarama.ProducerMessage{
		Topic:   topic,
		Key:     sarama.ByteEncoder(key),
		Value:   sarama.ByteEncoder(value),
		Headers: headers,
	}
	return p.producer.SendMessage(msg)
}

// SendMessageWithRequestID is a convenience wrapper that automatically adds a request_id header.
func (p *Producer) SendMessageWithRequestID(topic string, key, value []byte, requestID string) (partition int32, offset int64, err error) {
	headers := []sarama.RecordHeader{
		{Key: []byte("request_id"), Value: []byte(requestID)},
	}
	return p.SendMessage(topic, key, value, headers...)
}

// Close shuts down the producer.
func (p *Producer) Close() error {
	return p.producer.Close()
}
