package kafka

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/EbumbaE/bandit/pkg/logger"
)

type KafkaConsumer interface {
	Consume(ctx context.Context)
	Close() error
}

type Handler func(ctx context.Context, msg []byte) error

type kafkaConsumer struct {
	consumer   sarama.Consumer
	handler    Handler
	partitions []int32
	topic      string
}

func NewKafkaConsumer(ctx context.Context, brokers []string, topic string, handler Handler) (KafkaConsumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create consumer")
	}

	partitions, err := consumer.Partitions(topic)
	if err != nil {
		consumer.Close()
		return nil, errors.Wrap(err, "failed to get partitions")
	}

	return &kafkaConsumer{
		consumer:   consumer,
		topic:      topic,
		partitions: partitions,
		handler:    handler,
	}, nil
}

func (c *kafkaConsumer) Consume(ctx context.Context) {
	for _, partition := range c.partitions {
		partitionConsumer, err := c.consumer.ConsumePartition(c.topic, partition, sarama.OffsetNewest)
		if err != nil {
			logger.Error("failed to consume", zap.String("topic", c.topic), zap.Int32("partition", partition))
			return
		}
		defer partitionConsumer.Close()

		go func(pc sarama.PartitionConsumer) {
			for {
				select {
				case msg := <-pc.Messages():
					if err := c.handler(ctx, msg.Value); err != nil {
						logger.Error("message handling failed: ",
							zap.Int32("partition", msg.Partition),
							zap.Int64("offset", msg.Offset),
							zap.Error(err))
					}
				case err := <-pc.Errors():
					logger.Error("kafka consumer error: ",
						zap.Int32("partition", partition),
						zap.Error(err))
				case <-ctx.Done():
					return
				}
			}
		}(partitionConsumer)
	}

	<-ctx.Done()

	logger.Error("kafka consumer stop", zap.String("topic", c.topic))

	return
}

func (c *kafkaConsumer) Close() error {
	if err := c.consumer.Close(); err != nil {
		return errors.Wrap(err, "failed to close consumer")
	}
	return nil
}
