package kafka

import (
	"context"
	"log"

	"github.com/IBM/sarama"
)

type Producer struct {
	client sarama.SyncProducer
}

func NewProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Producer.Return.Successes = true

	client, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{client: client}, nil
}

func (p *Producer) Publish(_ context.Context, topic, key string, payload []byte) error {
	_, _, err := p.client.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(payload),
	})
	return err
}

func (p *Producer) Close() error {
	return p.client.Close()
}

type MessageHandler func(context.Context, *sarama.ConsumerMessage) error

type consumerGroupHandler struct {
	ctx     context.Context
	handler MessageHandler
}

func (c *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case <-c.ctx.Done():
			return nil
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			if err := c.handler(c.ctx, message); err != nil {
				log.Printf("consumer handler failed for topic=%s partition=%d offset=%d err=%v", message.Topic, message.Partition, message.Offset, err)
			}
			session.MarkMessage(message, "")
		}
	}
}

func ConsumeGroup(ctx context.Context, brokers []string, groupID string, topics []string, handler MessageHandler) error {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	group, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return err
	}
	defer func() {
		_ = group.Close()
	}()

	consumer := &consumerGroupHandler{
		ctx:     ctx,
		handler: handler,
	}

	for {
		if ctx.Err() != nil {
			return nil
		}

		if err = group.Consume(ctx, topics, consumer); err != nil {
			return err
		}
	}
}
