package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/ganlvtech/go-kahla-api/cryptojs"
	"github.com/ganlvtech/go-kahla-api/kahla"
	"github.com/ganlvtech/go-kahla-api/pusher"

	toast "github.com/ganlvtech/go-kahla-notify/snore-toast"
)

type Client struct {
	config     *Config
	client     *kahla.Client
	pusher     *pusher.Pusher
	snoreToast *toast.SnoreToast
}

func NewClient(config *Config) *Client {
	c := &Client{
		config: config,
		client: kahla.NewClient(config.ServerUrl, config.OssUrl),
	}
	if config.EnableSnoreToast {
		c.snoreToast = toast.New(config.SnoreToastPath)
	}
	return c
}

func (c *Client) downloadHeadImage(headImgFileKey uint32, filePath string) error {
	var w uint32 = 100
	var h uint32 = 100
	data, _, err := c.client.Oss.HeadImgFile(&kahla.Oss_Download_FromKeyRequest{
		HeadImgFileKey: headImgFileKey,
		W:              &w,
		H:              &h,
	})
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) toastAsync(title string, message string) {
	go c.snoreToast.Toast(title, message)
}

func (c *Client) toastWithHeadImageKey(title string, message string, headImgFileKey uint32) error {
	filePath := path.Join(c.config.AvatarsDir, fmt.Sprintf("%d.png", headImgFileKey))
	if fileExists(filePath) {
		return c.snoreToast.ToastWithImage(title, message, filePath)
	}

	done := make(chan bool)
	go func() {
		err := c.downloadHeadImage(headImgFileKey, filePath)
		if err != nil {
			log.Println("Download head image error:", err, "head image file key:", headImgFileKey, "file path:", filePath)
			done <- false
			return
		}
		err = c.snoreToast.ToastWithImage(title, message, filePath)
		if err != nil {
			log.Println("Toast with image error:", err, "file path:", filePath)
			done <- false
			return
		}
		done <- true
	}()
	select {
	case downloaded := <-done:
		if downloaded {
			err := c.snoreToast.ToastWithImage(title, message, filePath)
			if err == nil {
				return nil
			}
		}
	case <-time.After(10 * time.Second):
	}
	return c.snoreToast.Toast(title, message)
}

func (c *Client) toastWithHeadImageKeyAsync(title string, message string, headImgFileKey uint32) {
	go c.toastWithHeadImageKey(title, message, headImgFileKey)
}

func (c *Client) authByPassword() error {
	response, httpResponse, err := c.client.Auth.AuthByPassword(&kahla.Auth_AuthByPasswordRequest{
		Email:    c.config.Email,
		Password: c.config.Password,
	})
	if err != nil {
		return err
	}
	if httpResponse.StatusCode != http.StatusOK {
		return &HttpResponseStatusCodeNotOKError{httpResponse}
	}
	if response.Code != 0 {
		return &KahlaResponseCodeNotZeroError{
			Tag:     "AuthByPassword",
			Code:    response.Code,
			Message: response.Message,
		}
	}
	return nil
}

func (c *Client) initPusher() (*kahla.Auth_InitPusherResponse, error) {
	response, httpResponse, err := c.client.Auth.InitPusher()
	if err != nil {
		return response, err
	}
	if httpResponse.StatusCode != http.StatusOK {
		return response, &HttpResponseStatusCodeNotOKError{httpResponse}
	}
	if response.Code != 0 {
		return response, &KahlaResponseCodeNotZeroError{
			Tag:     "InitPusher",
			Code:    response.Code,
			Message: response.Message,
		}
	}
	return response, nil
}

func (c *Client) pusherDefaultEventHandler(i interface{}) {
	switch v := i.(type) {
	case *pusher.Pusher_NewMessageEvent:
		content, err := cryptojs.AesDecrypt(v.Content, v.AesKey)
		if err != nil {
			log.Println(err)
		} else {
			title := v.Sender.NickName
			message := content
			log.Println(title, ":", message)
		}
	case *pusher.Pusher_NewFriendRequestEvent:
		title := "Friend request"
		message := "You have got a new friend request!"
		log.Println(title, ":", message, "nick name:", v.Requester.NickName, "id:", v.Requester.Id)
	case *pusher.Pusher_WereDeletedEvent:
		title := "Were deleted"
		message := "You were deleted by one of your friends from his friend list."
		log.Println(title, ":", message, "nick name:", v.Trigger.NickName, "id:", v.Trigger.Id)
	case *pusher.Pusher_FriendAcceptedEvent:
		title := "Friend request"
		message := "Your friend request was accepted!"
		log.Println(title, ":", message, "nick name:", v.Target.NickName, "id:", v.Target.Id)
	case *pusher.Pusher_TimerUpdatedEvent:
		title := "Self-destruct timer updated!"
		message := fmt.Sprintf("Your current message life time is: %d", v.NewTimer)
		log.Println(title, ":", message, "conversation id:", v.ConversationId)
	case *pusher.Pusher_NewMemberEvent:
		title := "New member"
		message := fmt.Sprintf("%s has join the group.", v.NewMember.NickName)
		log.Println(title, ":", message, "conversation id:", v.ConversationId)
	case *pusher.Pusher_SomeoneLeftEvent:
		title := "Someone left"
		message := fmt.Sprintf("%s has successfully left the group.", v.LeftUser.NickName)
		log.Println(title, ":", message, "conversation id:", v.ConversationId)
	case *pusher.Pusher_DissolveEvent:
		title := "Group Dissolved"
		message := "A group dissolved"
		log.Println(title, ":", message, "conversation id:", v.ConversationId)
	}
}

func (c *Client) pusherToastEventHandler(i interface{}) {
	switch v := i.(type) {
	case *pusher.Pusher_NewMessageEvent:
		content, err := cryptojs.AesDecrypt(v.Content, v.AesKey)
		if err == nil {
			c.toastWithHeadImageKeyAsync(v.Sender.NickName, content, v.Sender.HeadImgFileKey)
		}
	case *pusher.Pusher_NewFriendRequestEvent:
		c.toastWithHeadImageKeyAsync("Friend request", fmt.Sprintf("You got a new friend request from %s.", v.Requester.NickName), v.Requester.HeadImgFileKey)
	case *pusher.Pusher_WereDeletedEvent:
		c.toastWithHeadImageKeyAsync("Were deleted", fmt.Sprintf("You were deleted by %s!", v.Trigger.NickName), v.Trigger.HeadImgFileKey)
	case *pusher.Pusher_FriendAcceptedEvent:
		c.toastWithHeadImageKeyAsync("Friend request", fmt.Sprintf("Your friend request to %s was accepted.", v.Target.NickName), v.Target.HeadImgFileKey)
	case *pusher.Pusher_TimerUpdatedEvent:
		// Do nothing
	case *pusher.Pusher_NewMemberEvent:
		// Do nothing
	case *pusher.Pusher_SomeoneLeftEvent:
		// Do nothing
	case *pusher.Pusher_DissolveEvent:
		c.toastAsync("Group dissolved", fmt.Sprintf("Group (id = %d) was dissolved!", v.ConversationId))
	default:
		log.Fatal("unknown event type")
	}
}

func (c *Client) Run(interrupt <-chan struct{}) error {
	err := c.authByPassword()
	if err != nil {
		return err
	}
	log.Println("Login OK.")
	response, err := c.initPusher()
	if err != nil {
		return err
	}
	log.Println("Init pusher OK.")
	c.pusher = pusher.New(response.ServerPath, c.pusherDefaultEventHandler)
	if c.config.EnableSnoreToast {
		c.pusher.EventHandlers = append(c.pusher.EventHandlers, c.pusherToastEventHandler)
	}
	log.Println("Connecting to pusher...")
	go func() {
		for v := range c.pusher.StateChangeChan {
			switch v {
			case pusher.WebSocketStateNew:
				panic("unreachable")
			case pusher.WebSocketStateConnected:
				log.Println("Connect to pusher OK.")
			case pusher.WebSocketStateDisconnected:
				log.Println("Disconnect from pusher.")
			case pusher.WebSocketStateClosed:
				log.Println("Pusher closed.")
			}
		}
	}()
	err = c.pusher.Connect(interrupt)
	if err != nil {
		log.Println("Disconnected from pusher.")
	}
	return err
}
