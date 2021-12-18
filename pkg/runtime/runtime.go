package runtime

import (
	"errors"
	"fmt"
	// "strings"
	"time"
	"encoding/json"

	log "github.com/sirupsen/logrus"

	"github.com/knwng/ppgi/pkg/algorithms/rsa_blind"
	"github.com/knwng/ppgi/pkg/graph"
)

type Intersecter interface {
	Run() error
	runClient() error
	runHost() error
}

type RSABlindRuntime struct {
	role 				string
	fetchInterval 		int
	intersect 			*rsa_blind.RSABlindIntersect
	producer 			Producer
	consumer 			Consumer
	kv 					KV
	graph				*graph.NebulaReadWriter
	nodes				[]graph.PrincipleNode
	lastGraphFetchTime 	time.Time
}

func NewRSABlindRuntime(role string, fetchInterval int,
		intersect *rsa_blind.RSABlindIntersect, producer Producer,
		consumer Consumer, kv KV, graph *graph.NebulaReadWriter,
		nodes []graph.PrincipleNode) (*RSABlindRuntime, error) {

	return &RSABlindRuntime{
		role: role,
		fetchInterval: fetchInterval,
		intersect: intersect,
		producer: producer,
		consumer: consumer,
		kv: kv,
		graph: graph,
		nodes: nodes,
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

func (s *RSABlindRuntime) lookupNewData() ([]string, time.Time, error) {
	totalData := make([]string, 0)

	current := time.Now()

	for _, node := range s.nodes {
		data, err := s.graph.LookupWithTimeLimit(&node, [2]time.Time{s.lastGraphFetchTime, current})
		if err != nil {
			return []string{}, time.Time{}, errors.New(fmt.Sprintf("Looking up node %+v in nebula graph failed, err: %s", node, err))
		}
		totalData = append(totalData, data...)
	}

	// if len(totalData) > 0 {
	// 	s.lastGraphFetchTime = current
	// }

	return totalData, current, nil
}

func (s *RSABlindRuntime) receiveMessage(c chan Message) {
	for {
		msg, err := s.consumer.ReceiveStruct()
		if err != nil {
			log.WithField("error", err).Warning("Client failed to receive message")
			continue
		}
		c <- msg
	}
}

func (s *RSABlindRuntime) runClient() error {
	// get data from consumer
	msgChan := make(chan Message)
	go s.receiveMessage(msgChan)

	fetchGraphTicker := time.NewTicker(time.Duration(s.fetchInterval) * time.Second)

	// client loop
	for {
		select {
		case msg := <-msgChan:
			// process received message
			if step := msg.Step; step == rsa_blind.StepHostSendPubKey {
				if len(msg.Key.N) == 0 || msg.Key.E <= 0 {
					log.WithField("msg", msg).Warning("Client received invalid key")
					continue
				}
				s.intersect.SetPubKey(msg.Key.N, msg.Key.E)
				// TODO(knwng): Think about recovery and retransmission, maybe add status & session to message?
			} else if step == rsa_blind.StepHostHash {
				// TODO(knwng): Need data input
			} else if step == rsa_blind.StepHostBlindSign {

			} else {
				log.WithField("msg", msg).Warning("Client received a message with wrong step")
				continue
			}
		case <-fetchGraphTicker.C:
			log.Info("Fetch principle node data from graph")

			// check whether key exchanging finished
			if !s.intersect.HasPubKey() {
				log.Warn("The client hasn't got pubkey yet, skip")
				continue
			}

			// fetch data
			data, newTime, err := s.lookupNewData()
			if err != nil {
				log.Errorf("Fetching data from graph database failed, err: %s", err)
				continue
			}

			if len(data) == 0 {
				log.WithFields(log.Fields{
					"start_time": s.lastGraphFetchTime.String(),
					"end_time": newTime,
				}).Info("No new data found in graph database")
				s.lastGraphFetchTime = newTime
				continue
			}

			yb, rands, err := s.intersect.ClientBlinding(data)
			if err != nil {
				log.WithField("data", data).Errorf("ClientBlinding failed, err: %s", err)
				continue
			}

			// send message to mq
			algorithm, step := "rsa", rsa_blind.StepClientBlind
			sessionKey := generateSessionKey(algorithm, step)

			if err = s.producer.SendStruct(&Message{
				Algorithm: algorithm,
				Step: step,
				SessionKey: sessionKey,
				Data: rsa_blind.BigIntsToBytesSlice(yb),
			}); err != nil {
				log.WithFields(log.Fields{
					"connection_info": s.producer.GetConnectionInfo(),
					"error": err,
				}).Warning("Sending message to mq failed")
			}

			// Send rands to kv
			encodedRands, err := json.Marshal(rands)
			if err != nil {
				log.WithFields(log.Fields{
					"rands": rands,
					"error": err,
				}).Error("Marshal rands to json failed")
				continue
			}

			if err = s.kv.Put(sessionKey, string(encodedRands)); err != nil {
				log.WithFields(log.Fields{
					"rands": rands,
					"session_key": sessionKey,
					"error": err,
				}).Error("Send rands to kv failed")
				continue
			}

			s.lastGraphFetchTime = newTime
		}
		
	}
}

func (s *RSABlindRuntime) runHost() error {
	// get data from consumer
	msgChan := make(chan Message)
	go s.receiveMessage(msgChan)

	fetchGraphTicker := time.NewTicker(time.Duration(s.fetchInterval) * time.Second)

	// host loop
	for {
		select {
		case <-fetchGraphTicker.C:
			// fetch data
			data, newTime, err := s.lookupNewData()
			if err != nil {
				log.Errorf("Fetching data from graph database failed, err: %s", err)
				continue
			}

			if len(data) == 0 {
				log.WithFields(log.Fields{
					"start_time": s.lastGraphFetchTime.String(),
					"end_time": newTime,
				}).Info("No new data found in graph database")
				s.lastGraphFetchTime = newTime
				continue
			}

			ta := s.intersect.HostOfflineHash(data)

			algorithm, step := "rsa", rsa_blind.StepHostHash

			sessionKey := generateSessionKey(algorithm, step)
			
			if err = s.producer.SendStruct(&Message{
				Algorithm: algorithm,
				Step: step,
				SessionKey: sessionKey,
				Data: ta,
			}); err != nil {
				log.WithFields(log.Fields{
					"connection_info": s.producer.GetConnectionInfo(),
					"error": err,
				}).Warning("Sending message to mq failed")
			}

			s.lastGraphFetchTime = newTime
		case msg := <-msgChan:
			if step := msg.Step; step == rsa_blind.StepClientReceivedPubKey {

			} else if step == rsa_blind.StepClientBlind {

			} else if step == rsa_blind.StepClientUnblind {

			} else {
				log.WithField("msg", msg).Warning("Host received a message with wrong step")
				continue
			}
		}
	}

}
