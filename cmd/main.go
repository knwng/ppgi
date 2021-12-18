package main

import (
	"os"
	"log"

	flags "github.com/jessevdk/go-flags"
	"github.com/spf13/viper"
	"github.com/knwng/ppgi/pkg/graph"
	"github.com/knwng/ppgi/pkg/runtime"
	"github.com/knwng/ppgi/pkg/algorithms/rsa_blind"
)

type Options struct {
	ConfigFile string `short:"c" long:"config" description:"config file" default:"ppgi.yaml"`
}

func main() {
	var options Options
	_, err := flags.Parse(&options)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(options.ConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	// load config file
	config := viper.New()
	config.SetConfigType("yaml")
	if err := config.ReadConfig(f); err != nil {
		log.Fatal(err)
	}

	// initialize kv
	var kv runtime.KV
	kvType := config.GetString("kv.type")
	switch kvType {
	case "redis":
		kv = runtime.NewRedisKV(config.GetString("kv.url"),
								config.GetString("kv.password"),
								config.GetInt("kv.db"))
	default:
		log.Fatalf("Unsupported kv type: %s", kvType)
	}

	// initialize mq producer and consumer
	var producer runtime.Producer
	var consumer runtime.Consumer
	mqURL := config.GetString("mq.url")
	mqInTopic := config.GetString("mq.in_topic")
	mqOutTopic := config.GetString("mq.out_topic")
	mqType := config.GetString("mq.type")

	switch mqType {
	case "pulsar":
		schema, err := runtime.ReadSchema(config.GetString("mq.schema"))
		if err != nil {
			log.Fatalf("Read schema file failed, err: %s", err)
		}
		if producer, err = runtime.NewPulsarProducer(mqURL, mqOutTopic, &schema); err != nil {
			log.Fatalf("Initialize pulsar producer failed, err: %s", err)
		}
		if consumer, err = runtime.NewPulsarConsumer(mqURL, mqInTopic, &schema); err != nil {
			log.Fatalf("Initialize pulsar consumer failed, err: %s", err)
		}
	default:
		log.Fatalf("Unsupported mq type: %s", mqType)
	}

	// initialize nebula graph client
	nebula, err := graph.NewNebulaReadWriter(config.GetString("graph.address"),
					config.GetInt("graph.port"),
					config.GetString("graph.username"),
					config.GetString("graph.password"),
					config.GetString("graph.graph_name"))
	if err != nil {
		log.Fatalf("Initializing nebula graph failed, err: %s", err)
	}

	// initialize runtime
	algorithmType := config.GetString("algorithm.type")
	role := config.GetString("role")
	nodes, err := graph.ParsePrincipleNodes(config.GetStringSlice("graph.principle_nodes"))
	if err != nil {
		log.Fatalf("Parsing principle node failed, err: %s", err)
	}

	var intersectRuntime runtime.Intersecter
	switch algorithmType {
	case "rsa":
		intersect, err := rsa_blind.NewRSABlindIntersect(
			config.GetInt("algorithm.key_bits"),
			config.GetString("algorithm.first_hash"),
			config.GetString("algorithm.second_hash"),
			role)
		if err != nil {
			log.Fatalf("Initialize RSA Intersection failed, err: %s", err)
		}
		interval := config.GetInt("graph.fetch_interval")
		intersectRuntime, err = runtime.NewRSABlindRuntime(role, interval,
			intersect, producer, consumer, kv, nebula, nodes)
		if err != nil {
			log.Fatalf("Initialize runtime failed, err: %s", err)
		}
	default:
		log.Fatalf("Unsupported algorithm: %s", algorithmType)
	}

	intersectRuntime.Run()
}
