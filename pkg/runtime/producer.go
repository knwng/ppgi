package runtime

import (
	"context"

	"github.com/apache/pulsar-client-go/pulsar"
)

type Producer interface {
	Send(payload []byte) error
	Close()
}

type pulsarProducer struct {
	client   pulsar.Client
	producer pulsar.Producer
}

func (p *pulsarProducer) Send(payload []byte) error {
	_, err := p.producer.Send(context.Background(), &pulsar.ProducerMessage{
		Payload: payload,
	})
	return err
}

func (p *pulsarProducer) Close() {
	p.producer.Close()
	p.client.Close()
}

func NewProducer(URL string, topic string) (Producer, error) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: URL,
	})
	if err != nil {
		return nil, err
	}

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic: topic,
	})
	if err != nil {
		return nil, err
	}

	return &pulsarProducer{
		client:   client,
		producer: producer,
	}, nil
}
