package main

import (
	"github.com/ganlvtech/go-kahla-notify/cryptojs"
	"github.com/ganlvtech/go-kahla-notify/kahla"
	"log"
)

func RunKahlaClient(email string, password string, webSocket *kahla.WebSocket, maxTryTimes int, done chan bool) {
	var err error
	for k1 := 0; k1 < maxTryTimes; k1++ {
		select {
		case <-done:
			log.Println("Main Kahla client worker stopped.")
			return
		default:
		}
		log.Println("Login as user:", email)
		client := kahla.NewClient()
		_, err = client.Auth.Login(email, password)
		if err != nil {
			log.Println("Login failed:", err)
			continue
		}
		log.Println("Login OK.")
		k1 = 0
		for k2 := 0; k2 < maxTryTimes; k2++ {
			select {
			case <-done:
				log.Println("Main Kahla client worker stopped.")
				return
			default:
			}
			log.Println("Initializing pusher.")
			var initPusherResponse *kahla.InitPusherResponse
			initPusherResponse, err = client.Auth.InitPusher()
			if err != nil {
				log.Println("Initialize pusher failed:", err)
				continue
			}
			log.Println("Initialize pusher OK. Server path:", initPusherResponse.ServerPath)
			k2 = 0
			for k3 := 0; k3 < maxTryTimes; k3++ {
				select {
				case <-done:
					log.Println("Main Kahla client worker stopped.")
					return
				default:
				}
				log.Println("Connecting to pusher.")
				err = webSocket.Connect(initPusherResponse.ServerPath)
				if err != nil {
					log.Println("Connect to pusher failed:", err)
					continue
				}
				log.Println("Connected to pusher OK.")
				k3 = 0
				select {
				case <-webSocket.Disconnected:
					log.Println("Disconnected.")
				case <-done:
					log.Println("Main Kahla client worker stopped.")
					return
				}
			}
		}
	}
}

func RunKahlaError(webSocket *kahla.WebSocket, done chan bool) {
	for {
		select {
		case <-done:
			log.Println("Error worker stopped.")
			return
		case err := <-webSocket.Err:
			log.Println(err)
		}
	}
}

func RunKahlaNotify(webSocket *kahla.WebSocket, done chan bool) {
	for {
		select {
		case <-done:
			log.Println("Notification worker stopped.")
			return
		case i := <-webSocket.Event:
			var title, message string
			switch v := i.(type) {
			case *kahla.NewMessageEvent:
				content, err := cryptojs.AesDecrypt(v.Content, v.AesKey)
				if err != nil {
					log.Println(err)
				} else {
					title = v.Sender.NickName+" [Kahla]"
					message = content
				}
			case *kahla.NewFriendRequestEvent:
				title = "Friend request"
				message = "You have got a new friend request!"
			case *kahla.WereDeletedEvent:
				title = "Were deleted"
				message = "You were deleted by one of your friends from his friend list."
			case *kahla.FriendAcceptedEvent:
				title = "Friend request"
				message = "Your friend request was accepted!"
			default:
				panic("invalid event type")
			}
			err := SnoreToast(title, message, "")
			if err != nil {
				log.Println(err)
			}
		}
	}
}
