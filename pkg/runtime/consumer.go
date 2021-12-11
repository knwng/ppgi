package runtime

import (
	"context"

	"github.com/apache/pulsar-client-go/pulsar"
)

type Consumer interface {
	Receive() ([]byte, error)
	Close()
}

type PulsarConsumer struct {
	client   pulsar.Client
	consumer pulsar.Consumer
}

func (c *PulsarConsumer) Receive() ([]byte, error) {
	msg, err := c.consumer.Receive(context.Background())
	return msg.Payload(), err
}

func (c *PulsarConsumer) Close() {
	c.consumer.Close()
	c.client.Close()
}

func NewPulsarConsumer(URL string, topic string) (Consumer, error) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: URL,
	})
	if err != nil {
		return nil, err
	}

	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: "my-sub",
		Type:             pulsar.Shared,
	})
	if err != nil {
		return nil, err
	}

	return &PulsarConsumer{
		client:   client,
		consumer: consumer,
	}, nil
}
