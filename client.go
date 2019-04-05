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
	Client     *kahla.Client
	Email      string
	Password   string
	ServerPath string
	WebSocket  *kahla.WebSocket
	SnoreToast *toast.SnoreToast
	AvatarsDir string
}

func NewClient(email string, password string, snoreToast *toast.SnoreToast, avatarsDir string) *Client {
	c := &Client{}
	c.Email = email
	c.Password = password
	c.Client = kahla.NewClient()
	c.WebSocket = kahla.NewWebSocket()
	c.SnoreToast = snoreToast
	c.AvatarsDir = avatarsDir
	return c
}

func (c *Client) toast(title string, message string) error {
	log.Println(title, ":", message)
	if c.SnoreToast != nil {
		return c.SnoreToast.Toast(title, message)
	}
	return nil
}

func (c *Client) toastWithImage(title string, message string, imagePath string) error {
	log.Println(title, ":", message)
	if c.SnoreToast != nil {
		return c.SnoreToast.ToastWithImage(title, message, imagePath)
	}
	return nil
}

func (c *Client) downloadHeadImage(headImgFileKey int, filePath string) error {
	data, err := c.Client.Oss.HeadImgFile(headImgFileKey, 100, 100)
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
	if c.SnoreToast != nil {
		filePath := path.Join(c.AvatarsDir, fmt.Sprintf("%d.png", headImgFileKey))
		if fileExists(filePath) {
			return c.toastWithImage(title, message, filePath)
		}

		go func() {
			err := c.downloadHeadImage(headImgFileKey, filePath)
			if err != nil {
				log.Println("Download head image error:", err, "head image file key:", headImgFileKey, "file path:", filePath)
			} else {
				err = c.toastWithImage(title, message, filePath)
				if err != nil {
					log.Println("Toast with image error:", err, "file path:", filePath)
				} else {
					return
				}
			}
			_ = c.toast(title, message)
		}()
		return nil
	}
	// Log message
	return c.toast(title, message)
}

func (c *Client) runNotifyUnread() {
	var response *kahla.MyFriendsResponse
	log.Println("Loading unread amount.")
	err := retry.Do(func() error {
		var err error
		response, err = c.Client.Friendship.MyFriends(false)
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
		case i := <-c.WebSocket.Event:
			switch v := i.(type) {
			case *kahla.NewMessageEvent:
				content, err := cryptojs.AesDecrypt(v.Content, v.AesKey)
				if err != nil {
					log.Println(err)
				} else {
					title := v.Sender.NickName + " [Kahla]"
					message := content
					headImgFileKey := v.Sender.HeadImgFileKey
					err = c.toastWithHeadImageKey(title, message, headImgFileKey)
					if err != nil {
						log.Println(err)
					}
				}
			case *kahla.NewFriendRequestEvent:
				title := "Friend request"
				message := "You have got a new friend request!"
				log.Println(title, ":", message)
				err := c.toast(title, message)
				if err != nil {
					log.Println(err)
				}
			case *kahla.WereDeletedEvent:
				title := "Were deleted"
				message := "You were deleted by one of your friends from his friend list."
				log.Println(title, ":", message)
				err := c.toast(title, message)
				if err != nil {
					log.Println(err)
				}
			case *kahla.FriendAcceptedEvent:
				title := "Friend request"
				message := "Your friend request was accepted!"
				log.Println(title, ":", message)
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
	log.Println("Login as user:", c.Email)
	err = retry.Do(func() error {
		_, err := c.Client.Auth.Login(c.Email, c.Password)
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
		response, err := c.Client.Auth.InitPusher()
		if err != nil {
			log.Println("Initialize pusher failed:", err, "Retry.")
			return err
		}
		c.ServerPath = response.ServerPath
		log.Println("Initialize pusher OK.")

		// Try connect to pusher
		log.Println("Connecting to pusher.")
		err = retry.Do(func() error {
			go func() {
				state := <-c.WebSocket.StateChanged
				if state == kahla.WebSocketStateConnected {
					log.Println("Connected to pusher OK.")
				}
			}()
			err := c.WebSocket.Connect(c.ServerPath, interrupt)
			if err != nil {
				if c.WebSocket.State == kahla.WebSocketStateClosed {
					log.Println("Keyboard interrupt:", err)
					return nil
				} else if c.WebSocket.State == kahla.WebSocketStateDisconnected {
					log.Println("Disconnected:", err, "Retry.")
					return err
				}
				log.Println("State:", c.WebSocket.State, "Error:", err, "Retry.")
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
