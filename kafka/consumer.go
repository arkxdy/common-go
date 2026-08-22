package kafka

import (
	"context"
	"log"

	"github.com/IBM/sarama"
)

// MessageHandler is a function that processes a single Kafka message.
// Return nil to mark the message as processed; return an error to skip marking.
type MessageHandler func(*sarama.ConsumerMessage) error

// Consumer wraps a sarama.ConsumerGroup and provides a simple, configurable consumer.
type Consumer struct {
	group       sarama.ConsumerGroup
	handler     MessageHandler
	errorLogger *log.Logger
	topics      []string
}

// NewConsumer creates a new consumer group from the given config.
func NewConsumer(cfg ConsumerConfig) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Version = cfg.Version
	config.Consumer.Offsets.Initial = cfg.InitialOffset
	config.Consumer.Return.Errors = true

	cg, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		group:       cg,
		handler:     cfg.Handler,
		errorLogger: cfg.ErrorLogger,
		topics:      cfg.Topics,
	}, nil
}

// Start begins consuming messages. It blocks until the context is cancelled.
// Errors during consumption are logged via the configured logger.
func (c *Consumer) Start(ctx context.Context) {
	if c.group == nil {
		c.errorLogger.Println("consumer group is nil, cannot start")
		return
	}

	handler := &consumerGroupHandler{
		handler:     c.handler,
		errorLogger: c.errorLogger,
	}

	for {
		select {
		case <-ctx.Done():
			c.errorLogger.Println("consumer context cancelled, shutting down")
			return
		default:
			if err := c.group.Consume(ctx, c.topics, handler); err != nil {
				c.errorLogger.Printf("consumer error: %v", err)
				// If context is cancelled, exit loop
				if ctx.Err() != nil {
					return
				}
			}
		}
	}
}

// Close closes the consumer group.
func (c *Consumer) Close() error {
	return c.group.Close()
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler.
type consumerGroupHandler struct {
	handler     MessageHandler
	errorLogger *log.Logger
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		if h.handler != nil {
			if err := h.handler(msg); err != nil {
				h.errorLogger.Printf("error processing message: %v", err)
				continue // do not mark, message will be redelivered
			}
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
