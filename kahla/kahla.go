package kahla

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
)

const KahlaServer string = "https://server.kahla.app"

type Client struct {
	client     http.Client
	Auth       *AuthService
	Friendship *FriendshipService
	Oss        *OssService
}

type service struct {
	client *Client
}

type AuthService service

type FriendshipService service

type OssService service

func NewClient() *Client {
	c := new(Client)
	c.client = http.Client{}
	c.client.Jar, _ = cookiejar.New(nil)
	c.Auth = &AuthService{c}
	c.Friendship = &FriendshipService{c}
	c.Oss = &OssService{c}
	return c
}

// func (c *Client) Do(req *http.Request, v interface{}) (*Response, error) {
// }

type ResponseStatusCodeNot200 struct {
	Response   *http.Response
	StatusCode int
}

func (r *ResponseStatusCodeNot200) Error() string {
	return fmt.Sprintf("response status code not 200: %d", r.StatusCode)
}

type ResponseCodeNotZero struct {
	Message string
}

func (r *ResponseCodeNotZero) Error() string {
	return "response code not zero: " + r.Message
}

type ResponseJsonDecodeError struct {
	Message string
	Err     error
}

func (r *ResponseJsonDecodeError) Error() string {
	if r.Message != "" {
		return r.Message
	}
	return r.Err.Error()
}
