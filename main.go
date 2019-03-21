package main

import (
	"github.com/ganlvtech/go-kahla-notify/kahla"
	"log"
	"os"
)

const ConfigFile = "config.json"

func main() {
	config, err := LoadConfigFromFile(ConfigFile)
	if err != nil {
		_, ok := err.(*os.PathError)
		if ok {
			err := SaveConfigToFile(ConfigFile, new(Config))
			if err != nil {
				panic(err)
			}
			log.Println("Please input your email and password in", ConfigFile)
			return
		}
		panic(err)
	}
	webSocket := kahla.NewWebSocket()
	done1 := make(chan bool)
	done2 := make(chan bool)
	done3 := make(chan bool)
	go RunKahlaError(webSocket, done1)
	go RunKahlaNotify(webSocket, done2)
	go RunKahlaClient(config.Email, config.Password, webSocket, 5, done3)
	WaitForCtrlC()
    select {
        case done1 <- true:
        default:
    }
    select {
        case done2 <- true:
        default:
    }
    select {
        case done3 <- true:
        default:
    }
	webSocket.Close()
}
