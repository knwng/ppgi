package main

import (
	"log"
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/spf13/viper"
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

	viper.SetConfigType("yaml")
	viper.ReadConfig(f)

	// algorithm := viper.GetString("algorithm")
	// kvType := viper.GetString("kv.type")
	// log.Printf("%s, %s", algorithm, kvType)
	
}
