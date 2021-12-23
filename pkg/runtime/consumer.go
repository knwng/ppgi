package runtime

import (
	"context"

	"github.com/apache/pulsar-client-go/pulsar"
)

type Consumer interface {
	Receive() ([]byte, error)
	ReceiveStruct() (Message, error)
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

func (c *PulsarConsumer) ReceiveStruct() (Message, error) {
	msg, err := c.consumer.Receive(context.Background())
	if err != nil {
		return Message{}, err
	}
	var s Message
	err = msg.GetSchemaValue(&s)
	return s, err
}

func (c *PulsarConsumer) Close() {
	c.consumer.Close()
	c.client.Close()
}

func NewPulsarConsumer(URL string, topic string, schema *string) (*PulsarConsumer, error) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: URL,
	})
	if err != nil {
		return nil, err
	}

	option := pulsar.ConsumerOptions{
		Topic: topic,
		SubscriptionName: "my-sub",
		Type: pulsar.Shared,
	}

	if schema != nil {
		option.Schema = pulsar.NewJSONSchema(*schema, nil)
	}

	consumer, err := client.Subscribe(option)
	if err != nil {
		return nil, err
	}

	return &PulsarConsumer{
		client:   client,
		consumer: consumer,
	}, nil
}
