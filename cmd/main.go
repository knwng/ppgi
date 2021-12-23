package main

import (
	"os"
	"path/filepath"
	"github.com/spf13/viper"
	log "github.com/sirupsen/logrus"
	flags "github.com/jessevdk/go-flags"

	"github.com/knwng/ppgi/pkg/graph"
	"github.com/knwng/ppgi/pkg/runtime"
	log_utils "github.com/knwng/ppgi/pkg/log"
	intersect_runtime "github.com/knwng/ppgi/pkg/intersect"
	"github.com/knwng/ppgi/pkg/algorithms/rsa_blind"
)

type Options struct {
	ConfigFile 	string 	`short:"c" long:"config" description:"config file" default:"ppgi.yaml"`
	Verbose		bool	`short:"v" long:"verbose" description:"Show verbose information"`
}

func init() {
}

func main() {
	var options Options
	_, err := flags.Parse(&options)
	checkErrOrFail(err)

	configFile, err := filepath.Abs(options.ConfigFile)
	checkErrOrFail(err)

	f, err := os.Open(configFile)
	checkErrOrFail(err)

	// load config file
	config := viper.New()
	config.SetConfigType("yaml")
	checkErrOrFail(config.ReadConfig(f))

	if logFile := config.GetString("log_file"); len(logFile) > 0 {
		checkErrOrFail(log_utils.SetLog(logFile, options.Verbose))
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
					config.GetString("graph.graph_name"),
					config.GetIntSlice("graph.neighbor_steps"))
	if err != nil {
		log.Fatalf("Initializing nebula graph failed, err: %s", err)
	}

	// initialize runtime
	algorithmType := config.GetString("algorithm.type")
	role := config.GetString("role")

	var intersectRuntime intersect_runtime.Intersecter
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
		timeout := config.GetInt("conn_timeout")
		graphDefinition := config.GetString("graph.graph_definition")
		intersectRuntime, err = intersect_runtime.NewRSABlindRuntime(role, interval,
			timeout, intersect, producer, consumer, kv, nebula, graphDefinition)
		if err != nil {
			log.Fatalf("Initialize runtime failed, err: %s", err)
		}
	default:
		log.Fatalf("Unsupported algorithm: %s", algorithmType)
	}

	intersectRuntime.Run()
}

func checkErrOrFail(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
