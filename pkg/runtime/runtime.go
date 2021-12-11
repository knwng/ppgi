package runtime

import (
	"fmt"
	"errors"

	"github.com/knwng/ppgi/pkg/algorithms/rsa_blind"
)

type Intersecter interface {
	Run() error
	runClient() error
	runHost() error
}

type RSABlindRuntime struct {
	role string
	intersect *rsa_blind.RSABlindIntersect
	producer Producer
	consumer Consumer
	kv KV
}

func NewRSABlindRuntime(role string, intersect *rsa_blind.RSABlindIntersect,
		producer Producer, consumer Consumer, kv KV) (*RSABlindRuntime, error) {

	return &RSABlindRuntime{
		role: role,
		intersect: intersect,
		producer: producer,
		consumer: consumer,
		kv: kv,
	}, nil
}

func (s *RSABlindRuntime) Run() error {
	if s.role == "client" {
		return s.runClient()
	} else if s.role == "host" {
		return s.runHost()
	} else {
		return errors.New(fmt.Sprintf("Unsupported role: %s", s.role))
	}
}

func (s *RSABlindRuntime) runClient() error {
	for {
		
	}
}

func (s *RSABlindRuntime) runHost() error {
	for {

	}
}
