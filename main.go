package main

import (
	"flag"
	toast "github.com/ganlvtech/go-kahla-notify/snore-toast"
	"log"
	"os"
	"os/signal"
)

const DefaultConfigFile = "config.json"

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

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
		err := SaveConfigToFile(configFile, new(Config))
		if err != nil {
			panic(err)
		}
		log.Println("Please input your email and password in:", configFile)
		return
	}

	config, err := LoadConfigFromFile(configFile)
	if err != nil {
		panic(err)
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	interrupt2 := make(chan struct{})
	go func() {
		<-interrupt
		close(interrupt2)
	}()

	var snoreToast *toast.SnoreToast
	if config.EnableSnoreToast {
		snoreToast = toast.New(config.SnoreToastPath)
	}

	c := NewClient(config.Email, config.Password, snoreToast, config.AvatarsDir)

	for {
		err = c.Run(interrupt2)
		if err != nil {
			log.Println(err)
		} else {
			break
		}
	}
}
