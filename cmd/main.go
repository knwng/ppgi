package main

import (
	"os"
	"log"

	flags "github.com/jessevdk/go-flags"
	"github.com/spf13/viper"
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
	switch config.GetString("kv.type") {
	case "redis":
		kv = runtime.NewRedisKV(config.GetString("kv.url"),
								config.GetString("kv.password"),
								config.GetInt("kv.db"))
	}

	// initialize mq producer and consumer
	var producer runtime.Producer
	var consumer runtime.Consumer
	mqURL := config.GetString("mq.url")
	mqInTopic := config.GetString("mq.in_topic")
	mqOutTopic := config.GetString("mq.out_topic")

	switch config.GetString("mq.type") {
	case "pulsar":
		if producer, err = runtime.NewPulsarProducer(mqURL, mqOutTopic); err != nil {
			log.Fatalf("Initialize pulsar producer failed, err: %s", err)
		}
		if consumer, err = runtime.NewPulsarConsumer(mqURL, mqInTopic); err != nil {
			log.Fatalf("Initialize pulsar consumer failed, err: %s", err)
		}
	}

	// initialize runtime
	algorithmType := config.GetString("algorithm.type")
	role := config.GetString("role")
	var intersectRuntime runtime.Intersecter
	switch algorithmType {
	case "rsa":
		intersect := rsa_blind.NewRSABlindIntersect(
			config.GetString("algorithm.first_hash"),
			config.GetString("algorithm.second_hash"))
		intersectRuntime, err = runtime.NewRSABlindRuntime(role, intersect,
												producer, consumer, kv)
		if err != nil {
			log.Fatalf("Initialize runtime failed, err: %s", err)
		}
	}

	intersectRuntime.Run()
}
