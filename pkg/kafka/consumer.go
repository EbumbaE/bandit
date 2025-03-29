package kafka

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/EbumbaE/bandit/pkg/logger"
)

type KafkaConsumer interface {
	Consume(ctx context.Context, handler func(msg []byte) error) error
	Close() error
}

type kafkaConsumer struct {
	consumer   sarama.Consumer
	partitions []int32
	topic      string
}

func NewKafkaConsumer(ctx context.Context, brokers []string, topic string) (KafkaConsumer, error) {
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
	}, nil
}

func (kc *kafkaConsumer) Consume(ctx context.Context, handler func(msg []byte) error) error {
	for _, partition := range kc.partitions {
		partitionConsumer, err := kc.consumer.ConsumePartition(kc.topic, partition, sarama.OffsetNewest)
		if err != nil {
			return errors.Wrapf(err, "failed to consume partition %d", partition)
		}
		defer partitionConsumer.Close()

		go func(pc sarama.PartitionConsumer) {
			for {
				select {
				case msg := <-pc.Messages():
					if err := handler(msg.Value); err != nil {
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
	return nil
}

func (kc *kafkaConsumer) Close() error {
	if err := kc.consumer.Close(); err != nil {
		return errors.Wrap(err, "failed to close consumer")
	}
	return nil
}
