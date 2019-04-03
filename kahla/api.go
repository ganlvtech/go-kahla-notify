package kahla

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type LoginResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

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
	Message    string `json:"message"`
}

func (s *AuthService) InitPusher() (*InitPusherResponse, error) {
	req, err := http.NewRequest("GET", KahlaServer + "/Auth/InitPusher", nil)
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

func (s *FriendshipService) MyFriends() (*MyFriendsResponse, error) {
	req, err := http.NewRequest("GET", KahlaServer+"/friendship/MyFriends?orderByName=false", nil)
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

func (s *OssService) HeadImgFile(headImgFileKey int) ([]byte, error) {
	resp, err := s.client.client.Get(fmt.Sprintf("https://oss.aiursoft.com/download/fromkey/%d?w=100&h=100", headImgFileKey))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, &ResponseStatusCodeNot200{Response: resp, StatusCode: resp.StatusCode}
	}
	return ioutil.ReadAll(resp.Body)
}
