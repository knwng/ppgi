package runtime

import (
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"

	"encoding/json"
	"time"

	"gopkg.in/yaml.v2"

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
	connTimeout			int
	algorithm			string
	intersect 			*rsa_blind.RSABlindIntersect
	producer 			Producer
	consumer 			Consumer
	kv 					KV
	graphClient			*graph.NebulaReadWriter
	graphDefinition			*graph.Graph
	// nodes				[]graph.PrincipleNode
	lastGraphFetchTime 	*time.Time
}

func NewRSABlindRuntime(role string, fetchInterval int, connTimeout int,
		intersect *rsa_blind.RSABlindIntersect, producer Producer,
		consumer Consumer, kv KV, graphClient *graph.NebulaReadWriter,
		graphDefinitionFn string) (*RSABlindRuntime, error) {

	data, err := ioutil.ReadFile(graphDefinitionFn)
	if err != nil {
		log.WithFields(log.Fields{
			"graph_definition_fn": graphDefinitionFn,
			"error": err,
		}).Fatal("Failed to read graph definition data")
	}

	graphDefinition := graph.Graph{}

	if err = yaml.Unmarshal(data, &graphDefinition); err != nil {
		log.Fatal("Failed to unmarshal graph definition")
	}

	reverseMap := make(map[string]*graph.Node)
	for _, node := range graphDefinition.Nodes {
		reverseMap[node.Type] = &node
	}

	graphDefinition.ReverseNodeMap = reverseMap

	return &RSABlindRuntime{
		role: role,
		fetchInterval: fetchInterval,
		connTimeout: connTimeout,
		algorithm: "rsa",
		intersect: intersect,
		producer: producer,
		consumer: consumer,
		kv: kv,
		graphClient: graphClient,
		graphDefinition: &graphDefinition,
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

	for _, node := range s.graphDefinition.Nodes {
		data, err := s.graphClient.LookupWithTimeLimit(&node, s.lastGraphFetchTime, &current)
		if err != nil {
			return []string{}, time.Time{}, errors.New(fmt.Sprintf("Faield to look up node %+v in nebula graph, err: %s", node, err))
		}
		totalData = append(totalData, data...)
	}

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

func (s *RSABlindRuntime) sendShutdown(sessionKey string) {
	if err := s.producer.SendStruct(&Message{
		Algorithm: s.algorithm,
		Step: rsa_blind.StepShutdown,
	}); err != nil {
		log.WithFields(log.Fields{
			"connection_info": s.producer.GetConnectionInfo(),
			"session_key": sessionKey,
			"error": err,
		}).Fatal("Failed to send shutdown message")
	}
}

func (s *RSABlindRuntime) sendMessageOrError(msg *Message) error {
	if err := s.producer.SendStruct(msg); err != nil {
		log.WithFields(log.Fields{
			"connection_info": s.producer.GetConnectionInfo(),
			"session_key": msg.SessionKey,
			"error": err,
		}).Error("Failed to send message to mq")
		return err
	}
	return nil
}

func (s *RSABlindRuntime) compareIDs() {

}

func (s *RSABlindRuntime) getRands(sessionKey string) ([]*big.Int, error) {
	randsStr, err := s.kv.HashGet("rand", sessionKey)
	if err != nil {
		log.WithField("error", err).Warning("Failed to get rands from kv")
		return []*big.Int{}, err
	}

	rands := make([]*big.Int, 0)

	if err := json.Unmarshal([]byte(randsStr), &rands); err != nil {
		log.WithFields(log.Fields{
			"rands_str": randsStr,
			"error": err,
		}).Warning("Failed to unmarshal rand message")
		return []*big.Int{}, err
	}

	return rands, nil
}

func (s *RSABlindRuntime) sendRands(sessionKey string, rands []*big.Int) error {
	encodedRands, err := json.Marshal(rands)
	if err != nil {
		log.WithField("error", err).Error("Failed to marshal rands to json")
		return err
	}

	if err = s.kv.HashPut("rand", map[string]string{sessionKey: string(encodedRands)}); err != nil {
		log.WithField("error", err).Error("Failed to send rands to kv")
		return err
	}

	return nil
}

func (s *RSABlindRuntime) delRands(sessionKey string) error {
	if err := s.kv.HashDel("rand", []string{sessionKey}); err != nil {
		log.WithFields(log.Fields{
			"session_key": sessionKey,
			"error": err,
		}).Error("Failed to delete used rands from kv")
		return err
	}
	return nil
}

func (s *RSABlindRuntime) sendOriginData(sessionKey string, data []string) error {
	encodedData, err := json.Marshal(data)
	if err != nil {
		log.WithField("error", err).Error("Failed to marshal data to json")
		return err
	}

	if err = s.kv.HashPut("origin_data", map[string]string{sessionKey: string(encodedData)}); err != nil {
		log.WithFields(log.Fields{
			"data": data,
			"session_key": sessionKey,
			"error": err,
		}).Error("Failed to send rands to kv")
		return err
	}

	return nil
}

func (s *RSABlindRuntime) getOriginData(sessionKey string) ([]string, error) {
	dataStr, err := s.kv.HashGet("origin_data", sessionKey);
	if err != nil {
		log.WithFields(log.Fields{
			"session_key": sessionKey,
			"error": err,
		}).Error("Failed to get origin data from kv")
		return []string{}, err
	}

	data := make([]string, 0)
	if err = json.Unmarshal([]byte(dataStr), &data); err != nil {
		log.WithFields(log.Fields{
			"data_str": dataStr,
			"error": err,
		}).Error("Failed to unmarshal original data")
		return []string{}, err
	}

	return data, nil
}

func (s *RSABlindRuntime) createAndSendHashIDMap(data []string, hash [][]byte) error {
	hashIDMap := make(map[string]string)

	if len(data) != len(hash) {
		err := errors.New("The sizes of original data and hash don't match")
		log.WithFields(log.Fields{
			"original_data": data,
			"hash": hash,
			"error": err,
		}).Error()
		return err
	}

	for i, ele := range data {
		hashIDMap[string(hash[i])] = ele
	}

	// send to kv
	if err := s.kv.HashPut("hash_id_map", hashIDMap); err != nil {
		log.WithField("error", err).Error("Failed to put HashIDMap to kv")
		return err
	}

	return nil
}

func (s *RSABlindRuntime) getMatchedId(hash []string) ([]string, error) {
	ret, err := s.kv.HashMultiGet("hash_id_map", hash)
	if err != nil {
		log.WithFields(log.Fields{
			"hash": hash,
			"error": err,
		}).Error("Failed to get matched id")
		return []string{}, err
	}

	data, _ := GetExistingStringAndIndex(ret)

	if len(data) == 0 {
		msg := "No hash matched"
		log.WithField("hash", hash).Warn(msg)
		return []string{}, errors.New(msg)
	}

	return data, nil	
}

func (s *RSABlindRuntime) sendMatchedId(data []string) error {
	if err := s.kv.SetAdd("matched_data", data); err != nil {
		log.WithFields(log.Fields{
			"data": data,
			"error": err,
		}).Error("Send matched id to kv set failed")
		return err
	}
	return nil
}

func getMapKeys(m map[string][]int) []string {
	j := 0
	rets := make([]string, len(m))
	for k := range m {
		rets[j] = k
		j++
	}
	return rets
}

func (s *RSABlindRuntime) matchIDAndSendData(msg *Message) error {
	hash := make([]string, len(msg.Data))
	for i, ele := range msg.Data {
		hash[i] = string(ele)
	}
	
	// get matched ids
	matchedID, err := s.getMatchedId(hash)
	if err != nil {
		return err
	}

	// add matched ids to kv set
	if err = s.sendMatchedId(matchedID); err != nil {
		return err
	}

	// get neighboring vertices and edges
	vertices, err := s.graphClient.GetAllNeighborVertices(matchedID)
	if err != nil {
		log.WithFields(log.Fields{
			"ids": matchedID,
			"error": err,
		}).Error("Failed to get neighboring vertices")
		return err
	}

	edges, err := s.graphClient.GetAllNeighborEdges(matchedID)
	if err != nil {
		log.WithFields(log.Fields{
			"ids": matchedID,
			"error": err,
		}).Error("Failed to get neighboring edges")
		return err
	}

	// filter non-matched vertices
	vertexVIDs := make([]string, len(vertices))
	j := 0
	for k := range vertices {
		vertexVIDs[j] = k
		j++
	}

	checkResult, err := s.kv.SetCheck("matched_data", vertexVIDs)
	if err != nil {
		return err
	}

	matchedVIDs := make(map[string]interface{})
	for i, flag := range checkResult {
		if flag {
			matchedVIDs[vertexVIDs[i]] = nil
		}
	}

	matchedVertices := make([]*graph.VertexData, 0)
	for k, v := range vertices {
		if _, ok := matchedVIDs[k]; ok {
			matchedVertices = append(matchedVertices, &v)
		}
	}

	matchedEdges := make([]*graph.EdgeData, 0)
	for _, e := range edges {
		_, srcOk := matchedVIDs[e.Source]
		_, dstOk := matchedVIDs[e.Destination]
		if srcOk && dstOk {
			matchedEdges = append(matchedEdges, &e)
		}
	}

	// send to host
	verticesEncoded, err := json.Marshal(matchedVertices)
	if err != nil {
		log.WithFields(log.Fields{
			"vertices": matchedVertices,
			"error": err,
		}).Error("Failed to marshal matched vertices to json")
		return err
	}

	edgesEncoded, err := json.Marshal(matchedEdges)
	if err != nil {
		log.WithFields(log.Fields{
			"edges": matchedEdges,
			"error": err,
		}).Error("Failed to marshal matched edges to json")
		return err
	}

	graphEncoded, err := json.Marshal(s.graphDefinition)
	if err != nil {
		log.WithFields(log.Fields{
			"graph_definition": s.graphDefinition,
			"error": err,
		}).Error("Failed to marshal matched graph definition to json")
		return err
	}

	if err := s.producer.SendStruct(&Message{
		Algorithm: s.algorithm,
		Step: rsa_blind.StepExchangeData,
		SessionKey: msg.SessionKey,
		Data: [][]byte{graphEncoded, verticesEncoded, edgesEncoded},
	}); err != nil {
		log.WithFields(log.Fields{
			"connection_info": s.producer.GetConnectionInfo(),
			"error": err,
		}).Error("Failed to send matched data through mq")
		return err
	}

	log.Info("Finished matching ids and sent matched data")
	return nil
}

func (s *RSABlindRuntime) loadDataToGraphDB(msg *Message) error {
	data := msg.Data
	if len(data) != 3 {
		log.WithField("message", msg).Error("The data field of StepExchangeData message has wrong format")
		return errors.New("Wrong data field")
	}

	_, verticesEncoded, edgesEncoded := data[0], data[1], data[2]
	vertices := make([]graph.VertexData, 0)
	edges := make([]graph.EdgeData, 0)

	if err := json.Unmarshal(verticesEncoded, &vertices); err != nil {
		log.WithFields(log.Fields{
			"encoded_vertices": verticesEncoded,
			"error": err,
		}).Error("Failed to unmarshal json-encoded vertices")
		return err
	}

	if err := json.Unmarshal(edgesEncoded, &edges); err != nil {
		log.WithFields(log.Fields{
			"encoded_edges": edgesEncoded,
			"error": err,
		}).Error("Failed to unmarshal json-encoded edges")
		return err
	}

	// add vertices and edges 
	// TODO(knwng): consider the situation when the definitions of two graphs are different
	if err := s.graphClient.AddVertexData(vertices); err != nil {
		return err
	}

	if err := s.graphClient.AddEdgeData(edges); err != nil {
		return err
	}

	log.Info("Data loaded to graph db")

	return nil
}

func (s *RSABlindRuntime) runClient() error {
	// get data from consumer
	msgChan := make(chan Message)
	go s.receiveMessage(msgChan)

	fetchGraphTicker := time.NewTicker(time.Duration(s.fetchInterval) * time.Second)
	defer fetchGraphTicker.Stop()

	log.Info("Waiting for incoming message")
	// client loop
	for {
		select {
		case <-fetchGraphTicker.C:
			log.Info("Fetch data from graph database periodically")
			// check whether key exchanging finished
			if !s.intersect.HasPubKey() {
				log.Warn("The client hasn't got pubkey yet, skip")
				continue
			}

			// fetch data
			data, newTime, err := s.lookupNewData()
			if err != nil {
				log.Errorf("Failed to fetch data from graph database, err: %s", err)
				continue
			}

			if len(data) == 0 {
				var lastTime string
				if s.lastGraphFetchTime == nil {
					lastTime = "no start time"
				} else {
					lastTime = s.lastGraphFetchTime.String()
				}
				log.WithFields(log.Fields{
					"start_time": lastTime,
					"end_time": newTime,
				}).Info("No new data found in graph database")
				s.lastGraphFetchTime = &newTime
				continue
			}

			yb, rands, err := s.intersect.ClientBlinding(data)
			if err != nil {
				log.WithField("data", data).Errorf("ClientBlinding failed, err: %s", err)
				continue
			}

			// send message to mq
			step := rsa_blind.StepClientBlind
			sessionKey := generateSessionKey(s.algorithm, step)

			if err = s.sendMessageOrError(&Message{
				Algorithm: s.algorithm,
				Step: step,
				SessionKey: sessionKey,
				Data: rsa_blind.BigIntsToBytesSlice(yb),
			}); err != nil {
				continue
			}

			// Send original data to kv
			if err = s.sendOriginData(sessionKey, data); err != nil {
				continue
			}

			// Send rands to kv
			if err = s.sendRands(sessionKey, rands); err != nil {
				continue
			}

			s.lastGraphFetchTime = &newTime
			log.Info("Client got new data from db, blind it, and send to host")
		case msg := <-msgChan:
			// process received message
			if step := msg.Step; step == rsa_blind.StepHostSendPubKey {
				log.Info("Client received pubkey from host")
				if s.intersect.HasPubKey() {
					log.Warning("Client has already had pubkey, skip")
				}

				if len(msg.Key.N) == 0 || msg.Key.E <= 0 {
					log.WithField("msg", msg).Warning("Client received invalid pubkey")
					s.sendShutdown(msg.SessionKey)
					continue
				}

				s.intersect.SetPubKey(msg.Key.N, msg.Key.E)

				// send ack message
				s.sendMessageOrError(&Message{
					Algorithm: s.algorithm,
					Step: rsa_blind.StepClientRcvPubKey,
					SessionKey: msg.SessionKey,
				})
			} else if step == rsa_blind.StepHostBlindSign {
				log.Info("Client starts to unblind signs from host")

				// get rands from kv
				rands, err := s.getRands(msg.SessionKey)
				if err != nil {
					continue
				}

				tb := s.intersect.ClientUnblinding(rsa_blind.BytesSliceToBigInts(msg.Data), rands)

				s.sendMessageOrError(&Message{
					Algorithm: s.algorithm,
					Step: rsa_blind.StepClientUnblind,
					SessionKey: msg.SessionKey,
					Data: tb,
				})

				// get origin data
				data, err := s.getOriginData(msg.SessionKey)
				if err != nil {
					continue
				}

				// combine hash and data, and send to kv
				if err = s.createAndSendHashIDMap(data, tb); err != nil {
					continue
				}

				// delete rands
				s.delRands(msg.SessionKey)
				log.Info("Client unblind the sign from host and send the hash to host")
			} else if step == rsa_blind.StepHostHash {
				// compare hash with current ID
				log.Info("Client starts to compare hash from host")
				if err := s.matchIDAndSendData(&msg); err != nil {
					continue
				}
			} else if step == rsa_blind.StepExchangeData {
				// load data to nebula graph
				if err := s.loadDataToGraphDB(&msg); err != nil {
					continue
				}

			} else {
				log.WithField("msg", msg).Warning("Client received a message with wrong step")
				continue
			}
		}
		
	}
}

func (s *RSABlindRuntime) pubKeyExchange() {
	log.Info("Host send pubkey to client")
	step := rsa_blind.StepHostSendPubKey
	sessionKey := generateSessionKey(s.algorithm, step)
	n, e := s.intersect.GetPubKey()
	key := Key{
		N: n,
		E: e,
	}

	// TODO(knwng): need to retry here
	if err := s.producer.SendStruct(&Message{
		Algorithm: s.algorithm,
		Step: step,
		SessionKey: sessionKey,
		Key: key,
	}); err != nil {
		log.WithField("error", err).Fatal("Host failed to send pubkey")
	}

	// receive message from consumer
	msgChan := make(chan Message)
	go s.receiveMessage(msgChan)

	log.Info("Host's waiting for ack of pubkey from client")
	for {
		select {
		case msg := <-msgChan:
			if msg.Step != rsa_blind.StepClientRcvPubKey {
				continue
			} else {
				return
			}
		case <-time.After(time.Duration(s.connTimeout) * time.Second):
			log.WithField("session_key", sessionKey).
				Fatal("Host didn't receive client's ack after sending pubkey")
		}
	}
}

func (s *RSABlindRuntime) runHost() error {
	// pubkey exchange
	s.pubKeyExchange()

	// receive message from consumer
	msgChan := make(chan Message)
	go s.receiveMessage(msgChan)

	fetchGraphTicker := time.NewTicker(time.Duration(s.fetchInterval) * time.Second)
	defer fetchGraphTicker.Stop()

	log.Info("Waiting for incoming message")
	// host loop
	for {
		select {
		case <-fetchGraphTicker.C:
			log.Info("Fetch data from graph database periodically")
			data, newTime, err := s.lookupNewData()
			if err != nil {
				log.Errorf("Failed to fetch data from graph database, err: %s", err)
				continue
			}

			if len(data) == 0 {
				var lastTime string
				if s.lastGraphFetchTime == nil {
					lastTime = "no start time"
				} else {
					lastTime = s.lastGraphFetchTime.String()
				}
				log.WithFields(log.Fields{
					"start_time": lastTime,
					"end_time": newTime,
				}).Info("No new data found in graph database")
				s.lastGraphFetchTime = &newTime
				continue
			}

			ta := s.intersect.HostOfflineHash(data)

			// send hash-data map to kv
			hashDataMap := make(map[string]string)
			for i, hash := range ta {
				hashDataMap[string(hash)] = data[i]
			}
			if err = s.kv.HashPut("hash_id_map", hashDataMap); err != nil {
				log.WithFields(log.Fields{
					"hash_data_map": hashDataMap,
					"error": err,
				}).Error("Failed to send hash-data map to kv")
				continue
			}

			step := rsa_blind.StepHostHash

			sessionKey := generateSessionKey(s.algorithm, step)

			if err = s.sendMessageOrError(&Message{
				Algorithm: s.algorithm,
				Step: step,
				SessionKey: sessionKey,
				Data: ta,
			}); err != nil {
				continue
			}

			s.lastGraphFetchTime = &newTime
			log.Info("Host got data from graph db, calculated hash and sent it to client")
		case msg := <-msgChan:
			// process message received from mq
			if step := msg.Step; step == rsa_blind.StepClientBlind {
				log.Info("Host starts to blind sign hash from client")
				zb := s.intersect.HostBlindSigning(rsa_blind.BytesSliceToBigInts(msg.Data))
				s.sendMessageOrError(&Message{
					Algorithm: s.algorithm,
					Step: rsa_blind.StepHostBlindSign,
					SessionKey: msg.SessionKey,
					Data: rsa_blind.BigIntsToBytesSlice(zb),
				})
			} else if step == rsa_blind.StepClientUnblind {
				// compare hash with current ID
				log.Info("Host starts to compare hash from client")
				if err := s.matchIDAndSendData(&msg); err != nil {
					continue
				}

			} else if step == rsa_blind.StepClientRcvPubKey {
				// consume reluctant pubkey ack message
				log.Info("Host received pubkey ack from client after key exchange, skip")
			} else if step == rsa_blind.StepExchangeData {
				// load data to nebula graph
				if err := s.loadDataToGraphDB(&msg); err != nil {
					continue
				}

			} else {
				log.WithField("msg", msg).Warning("Host received a message with wrong step")
				continue
			}
		}
	}

}

func bytesSliceToStringSlice(data [][]byte) []string {
	ret := make([]string, len(data))
	for i, ele := range data {
		ret[i] = string(ele)
	}
	return ret
}
