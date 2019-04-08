package main

import (
	"fmt"
	"github.com/avast/retry-go"
	"github.com/ganlvtech/go-kahla-notify/cryptojs"
	"github.com/ganlvtech/go-kahla-notify/kahla"
	"github.com/ganlvtech/go-kahla-notify/snore-toast"
	"io/ioutil"
	"log"
	"path"
)

type Client struct {
	client     *kahla.Client
	email      string
	password   string
	serverPath string
	webSocket  *kahla.WebSocket
	snoreToast *toast.SnoreToast
	avatarsDir string
}

func NewClient(email string, password string, snoreToast *toast.SnoreToast, avatarsDir string) *Client {
	c := &Client{}
	c.email = email
	c.password = password
	c.client = kahla.NewClient()
	c.webSocket = kahla.NewWebSocket()
	c.snoreToast = snoreToast
	c.avatarsDir = avatarsDir
	return c
}

func (c *Client) toast(title string, message string) error {
	if c.snoreToast != nil {
		return c.snoreToast.Toast(title, message)
	}
	return nil
}

func (c *Client) toastWithImage(title string, message string, imagePath string) error {
	if c.snoreToast != nil {
		return c.snoreToast.ToastWithImage(title, message, imagePath)
	}
	return nil
}

func (c *Client) downloadHeadImage(headImgFileKey int, filePath string) error {
	data, err := c.client.Oss.HeadImgFile(headImgFileKey, 100, 100)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) toastWithHeadImageKey(title string, message string, headImgFileKey int) error {
	if c.snoreToast != nil {
		filePath := path.Join(c.avatarsDir, fmt.Sprintf("%d.png", headImgFileKey))
		if fileExists(filePath) {
			return c.toastWithImage(title, message, filePath)
		}

		go func() {
			err := c.downloadHeadImage(headImgFileKey, filePath)
			if err != nil {
				log.Println("Download head image error:", err, "head image file key:", headImgFileKey, "file path:", filePath)
				_ = c.toast(title, message)
				return
			}
			err = c.toastWithImage(title, message, filePath)
			if err != nil {
				log.Println("Toast with image error:", err, "file path:", filePath)
				_ = c.toast(title, message)
				return
			}
		}()
		return nil
	}
	return nil
}

func (c *Client) runNotifyUnread() {
	var response *kahla.MyFriendsResponse
	log.Println("Loading unread amount.")
	err := retry.Do(func() error {
		var err error
		response, err = c.client.Friendship.MyFriends(false)
		if err != nil {
			log.Println("Loading unread amount failed:", err, "Retry.")
			return err
		}
		return nil
	})
	if err != nil {
		log.Println("Loading unread amount failed too many times:", err)
		return
	}

	for _, v := range response.Items {
		if v.UnReadAmount > 0 {
			message, err := cryptojs.AesDecrypt(v.LatestMessage, v.AesKey)
			if err != nil {
				log.Println("crypto.js AES decode error:", err)
				message = v.LatestMessage
			}
			title := fmt.Sprintf("[%d unread] ", v.UnReadAmount) + v.DisplayName + " [Kahla]"
			log.Println(title, ":", message)
			headImgFileKey := v.DisplayImageKey
			err = c.toastWithHeadImageKey(title, message, headImgFileKey)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (c *Client) runNotify(interrupt chan struct{}) {
	for {
		select {
		case <-interrupt:
			log.Println("Notification worker stopped.")
			return
		case i := <-c.webSocket.Event:
			switch v := i.(type) {
			case *kahla.NewMessageEvent:
				content, err := cryptojs.AesDecrypt(v.Content, v.AesKey)
				if err != nil {
					log.Println(err)
				} else {
					title := v.Sender.NickName + " [Kahla]"
					message := content
					log.Println(title, ":", message)
					headImgFileKey := v.Sender.HeadImgFileKey
					err := c.toastWithHeadImageKey(title, message, headImgFileKey)
					if err != nil {
						log.Println(err)
					}
				}
			case *kahla.NewFriendRequestEvent:
				title := "Friend request"
				message := "You have got a new friend request!"
				log.Println(title, ":", message, "nick name:", v.Requester.NickName, "id:", v.Requester.ID)
				err := c.toast(title, message)
				if err != nil {
					log.Println(err)
				}
			case *kahla.WereDeletedEvent:
				title := "Were deleted"
				message := "You were deleted by one of your friends from his friend list."
				log.Println(title, ":", message, "nick name:", v.Trigger.NickName, "id:", v.Trigger.ID)
				err := c.toast(title, message)
				if err != nil {
					log.Println(err)
				}
			case *kahla.FriendAcceptedEvent:
				title := "Friend request"
				message := "Your friend request was accepted!"
				log.Println(title, ":", message, "nick name:", v.Target.NickName, "id:", v.Target.ID)
				err := c.toast(title, message)
				if err != nil {
					log.Println(err)
				}
			case *kahla.TimerUpdatedEvent:
				title := "Self-destruct timer updated!"
				message := fmt.Sprintf("Your current message life time is: %d", v.NewTimer)
				log.Println(title, ":", message, "conversation id:", v.ConversationID)
				err := c.toast(title, message)
				if err != nil {
					log.Println(err)
				}
			default:
				panic("invalid event type")
			}
		}
	}
}

func (c *Client) Run(interrupt chan struct{}) error {
	var err error

	// Try login
	log.Println("Login as user:", c.email)
	err = retry.Do(func() error {
		_, err := c.client.Auth.Login(c.email, c.password)
		if err != nil {
			log.Println("Login failed:", err, "Retry.")
			return err
		}
		return nil
	})
	if err != nil {
		log.Println("Login failed too many times:", err)
		return err
	}
	log.Println("Login OK.")

	// Unread amount
	go c.runNotifyUnread()

	interrupt2 := make(chan struct{})
	defer close(interrupt2)
	go c.runNotify(interrupt2)

	// Try connect to pusher
	log.Println("Initializing pusher.")
	err = retry.Do(func() error {
		// Try initialize pusher
		response, err := c.client.Auth.InitPusher()
		if err != nil {
			log.Println("Initialize pusher failed:", err, "Retry.")
			return err
		}
		c.serverPath = response.ServerPath
		log.Println("Initialize pusher OK.")

		// Try connect to pusher
		log.Println("Connecting to pusher.")
		err = retry.Do(func() error {
			go func() {
				state := <-c.webSocket.StateChanged
				if state == kahla.WebSocketStateConnected {
					log.Println("Connected to pusher OK.")
				}
			}()
			err := c.webSocket.Connect(c.serverPath, interrupt)
			if err != nil {
				if c.webSocket.State == kahla.WebSocketStateClosed {
					log.Println("Keyboard interrupt:", err)
					return nil
				} else if c.webSocket.State == kahla.WebSocketStateDisconnected {
					log.Println("Disconnected:", err, "Retry.")
					return err
				}
				log.Println("State:", c.webSocket.State, "Error:", err, "Retry.")
				return err
			}
			log.Println("Keyboard interrupt.")
			return nil
		})
		if err != nil {
			log.Println("Connected to pusher failed too many times:", err)
			return err
		}
		return nil
	})
	if err != nil {
		log.Println("Initialize pusher failed too many times:", err)
		return err
	}
	log.Println("Kahla client stopped.")
	return nil
}
