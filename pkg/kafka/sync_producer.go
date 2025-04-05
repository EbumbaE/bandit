package kafka

import (
	"context"

	"github.com/IBM/sarama"
	"github.com/pkg/errors"
)

type SyncProducer interface {
	SendMessage(ctx context.Context, key, value []byte) error
	Close() error
}

type syncProducer struct {
	topic    string
	producer sarama.SyncProducer
}

func NewSyncProducer(ctx context.Context, topic string, brokers []string) (SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	pr, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create producer")
	}

	return &syncProducer{
		producer: pr,
		topic:    topic,
	}, nil
}

func (p *syncProducer) SendMessage(ctx context.Context, key, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	}

	_, _, err := p.producer.SendMessage(msg)
	if err != nil {
		return errors.Wrapf(err, "failed to send message: topic[%s]", p.topic)
	}
	return nil
}

func (p *syncProducer) Close() error {
	if err := p.producer.Close(); err != nil {
		return errors.Wrap(err, "failed to close sync producer")
	}
	return nil
}
