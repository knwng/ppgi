package runtime

import (
	"fmt"
	"context"

	"github.com/apache/pulsar-client-go/pulsar"
)

type Producer interface {
	Send(payload []byte) error
	SendStruct(msg *Message) error
	GetConnectionInfo() string
	Close()
}

type PulsarProducer struct {
	client   	pulsar.Client
	producer 	pulsar.Producer
	url			string
	topic		string
}

func (p *PulsarProducer) GetConnectionInfo() string {
	return fmt.Sprintf("url: %s, topic: %s", p.url, p.topic)
}

func (p *PulsarProducer) Send(payload []byte) error {
	_, err := p.producer.Send(context.Background(), &pulsar.ProducerMessage{
		Payload: payload,
	})
	return err
}

func (p *PulsarProducer) SendStruct(msg *Message) error {
	_, err := p.producer.Send(context.Background(), &pulsar.ProducerMessage{
		Value: msg,
	})
	return err
}

func (p *PulsarProducer) Close() {
	p.producer.Close()
	p.client.Close()
}

func NewPulsarProducer(URL string, topic string, schema *string) (*PulsarProducer, error) {
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: URL,
	})
	if err != nil {
		return nil, err
	}

	option := pulsar.ProducerOptions{
		Topic: topic,
	}

	if schema != nil {
		option.Schema = pulsar.NewJSONSchema(*schema, nil)
	}

	producer, err := client.CreateProducer(option)

	if err != nil {
		return nil, err
	}

	return &PulsarProducer{
		client:   client,
		producer: producer,
	}, nil
}
