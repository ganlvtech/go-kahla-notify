package kahla

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type User struct {
	MakeEmailPublic   bool        `json:"makeEmailPublic"`
	Email             string      `json:"email"`
	ID                string      `json:"id"`
	Bio               string      `json:"bio"`
	NickName          string      `json:"nickName"`
	Sex               interface{} `json:"sex"`
	HeadImgFileKey    int         `json:"headImgFileKey"`
	PreferedLanguage  string      `json:"preferedLanguage"`
	AccountCreateTime string      `json:"accountCreateTime"`
	EmailConfirmed    bool        `json:"emailConfirmed"`
}

type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// POST https://server.kahla.app/Auth/AuthByPassword
//
// Email=123@abc.com&Password=123456
func (s *AuthService) Login(email string, password string) (*LoginResponse, error) {
	v := url.Values{}
	v.Add("Email", email)
	v.Add("Password", password)
	req, err := NewPostRequest(KahlaServer+"/Auth/AuthByPassword", v)
	if err != nil {
		return nil, err
	}
	response := &LoginResponse{}
	_, err = s.client.Do(req, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

type InitPusherResponse struct {
	ServerPath string `json:"serverPath"`
	ChannelID  int    `json:"channelId"`
	ConnectKey string `json:"connectKey"`
	Code       int    `json:"code"`
	Message    string `json:" "`
}

// GET https://server.kahla.app/Auth/InitPusher
func (s *AuthService) InitPusher() (*InitPusherResponse, error) {
	req, err := http.NewRequest("GET", KahlaServer+"/Auth/InitPusher", nil)
	if err != nil {
		return nil, err
	}
	response := &InitPusherResponse{}
	_, err = s.client.Do(req, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

type MyFriendsResponse struct {
	Items []struct {
		DisplayName       string    `json:"displayName"`
		DisplayImageKey   int       `json:"displayImageKey"`
		LatestMessage     string    `json:"latestMessage"`
		LatestMessageTime time.Time `json:"latestMessageTime"`
		UnReadAmount      int       `json:"unReadAmount"`
		ConversationID    int       `json:"conversationId"`
		Discriminator     string    `json:"discriminator"`
		UserID            string    `json:"userId"`
		AesKey            string    `json:"aesKey"`
		Muted             bool      `json:"muted"`
	} `json:"items"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// GET https://server.kahla.app/friendship/MyFriends?orderByName=false
func (s *FriendshipService) MyFriends(orderByName bool) (*MyFriendsResponse, error) {
	v := url.Values{}
	v.Set("orderByName", strconv.FormatBool(orderByName))
	req, err := http.NewRequest("GET", KahlaServer+"/friendship/MyFriends?"+v.Encode(), nil)
	if err != nil {
		return nil, err
	}
	response := &MyFriendsResponse{}
	_, err = s.client.Do(req, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

type MyRequestsResponse struct {
	Items []struct {
		ID         int       `json:"id"`
		CreatorID  string    `json:"creatorId"`
		Creator    User      `json:"creator"`
		TargetID   string    `json:"targetId"`
		CreateTime time.Time `json:"createTime"`
		Completed  bool      `json:"completed"`
	} `json:"items"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// GET https://server.kahla.app/friendship/MyRequests
func (s *FriendshipService) MyRequests() (*MyRequestsResponse, error) {
	req, err := http.NewRequest("GET", KahlaServer+"/friendship/MyRequests", nil)
	if err != nil {
		return nil, err
	}
	response := &MyRequestsResponse{}
	_, err = s.client.Do(req, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

type CompleteRequestResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// POST https://server.kahla.app/friendship/CompleteRequest/139
// accept=true
func (s *FriendshipService) CompleteRequest(requestId int, accept bool) (*CompleteRequestResponse, error) {
	v := url.Values{}
	v.Add("accept", strconv.FormatBool(accept))
	req, err := NewPostRequest(KahlaServer+"/Friendship/CompleteRequest/"+strconv.Itoa(requestId), v)
	if err != nil {
		return nil, err
	}
	response := &CompleteRequestResponse{}
	_, err = s.client.Do(req, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// GET https://oss.aiursoft.com/download/fromkey/2611?w=100&h=100
func (s *OssService) HeadImgFile(headImgFileKey int, w int, h int) ([]byte, error) {
	v := url.Values{}
	v.Set("w", strconv.Itoa(w))
	v.Set("h", strconv.Itoa(h))
	resp, err := s.client.client.Get("https://oss.aiursoft.com/download/fromkey/" + strconv.Itoa(headImgFileKey) + "?" + v.Encode())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, &ResponseStatusCodeNot200{Response: resp, StatusCode: resp.StatusCode}
	}
	return ioutil.ReadAll(resp.Body)
}

type SendMessageResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// POST https://server.kahla.app/conversation/SendMessage/68
//
// content=content
func (c *ConversationService) SendMessage(conversationId int, content string) (*SendMessageResponse, error) {
	v := url.Values{}
	v.Add("content", content)
	req, err := NewPostRequest(KahlaServer+"/conversation/SendMessage/"+strconv.Itoa(conversationId), v)
	if err != nil {
		return nil, err
	}
	response := &SendMessageResponse{}
	_, err = c.client.Do(req, response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
