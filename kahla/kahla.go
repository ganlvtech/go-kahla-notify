package kahla

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"reflect"
	"strings"
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

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewClient() *Client {
	c := new(Client)
	c.client = http.Client{}
	c.client.Jar, _ = cookiejar.New(nil)
	c.Auth = &AuthService{c}
	c.Friendship = &FriendshipService{c}
	c.Oss = &OssService{c}
	return c
}

type ResponseStatusCodeNot200 struct {
	Response   *http.Response
	StatusCode int
}

func (r *ResponseStatusCodeNot200) Error() string {
	return fmt.Sprintf("response status code not 200: %d", r.StatusCode)
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

type ResponseCodeNotZero struct {
	Message string
}

func (r *ResponseCodeNotZero) Error() string {
	return "response code not zero: " + r.Message
}

func initializeResponse(i interface{}) {
	v := reflect.ValueOf(i)
	v = v.Elem()
	v.FieldByName("Code").SetInt(-1)
}

func castToResponse(i interface{}) *Response {
	v := reflect.ValueOf(i)
	v = v.Elem()
	response := &Response{}
	response.Message = v.FieldByName("Message").String()
	response.Code = int(v.FieldByName("Code").Int())
	return response
}

// do http response
//
// v must be a pointer to Response struct, which contains Message and Code field.
// json data returned via v.
func NewPostRequest(url string, data url.Values) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return req, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

// do http response
//
// v must be a pointer to Response struct, which contains Message and Code field.
// json data returned via v.
func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()
	if err != nil {
		return resp, err
	}
	if resp.StatusCode != 200 {
		return resp, &ResponseStatusCodeNot200{Response: resp, StatusCode: resp.StatusCode}
	}
	initializeResponse(v)
	err = json.NewDecoder(resp.Body).Decode(v)
	response := castToResponse(v)
	if err != nil {
		return resp, &ResponseJsonDecodeError{Message: response.Message, Err: err}
	}
	if response.Code != 0 {
		return resp, &ResponseCodeNotZero{response.Message}
	}
	return resp, nil
}
