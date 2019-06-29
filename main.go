package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/avast/retry-go"
)

const DefaultConfigFile = "config.json"

func main() {
	var configFile string
	var h bool
	var help bool
	flag.StringVar(&configFile, "config", DefaultConfigFile, "config path")
	flag.BoolVar(&h, "h", false, "help")
	flag.BoolVar(&help, "help", false, "help")
	flag.Parse()
	if h || help {
		flag.PrintDefaults()
		return
	}

	if !fileExists(configFile) {
		err := SaveConfigToFile(configFile, &Config{})
		if err != nil {
			log.Fatal(err)
		}
		log.Println("Please input your email and password in:", configFile)
		return
	}

	config, err := LoadConfigFromFile(configFile)
	if err != nil {
		log.Fatal(err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	interrupt2 := make(chan struct{})
	go func() {
		<-interrupt
		close(interrupt2)
	}()

	c := NewClient(config)
	err = retry.Do(func() error {
		return c.Run(interrupt2)
	})
	if err != nil {
		log.Println(err)
	} else {
		log.Println("Interrupt")
	}
}
