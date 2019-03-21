package kahla

import (
	"encoding/json"
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
	resp, err := s.client.client.PostForm(KahlaServer+"/Auth/AuthByPassword", v)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, &ResponseStatusCodeNot200{Response: resp, StatusCode: resp.StatusCode}
	}
	j := json.NewDecoder(resp.Body)
	response := &LoginResponse{Code: -1}
	err = j.Decode(response)
	if err != nil {
		return response, &ResponseJsonDecodeError{response.Message, err}
	}
	if response.Code != 0 {
		return response, &ResponseCodeNotZero{response.Message}
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
	resp, err := s.client.client.Get(KahlaServer + "/Auth/InitPusher")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, &ResponseStatusCodeNot200{Response: resp, StatusCode: resp.StatusCode}
	}
	j := json.NewDecoder(resp.Body)
	response := &InitPusherResponse{Code: -1}
	err = j.Decode(response)
	if err != nil {
		return response, &ResponseJsonDecodeError{response.Message, err}
	}
	if response.Code != 0 {
		return response, &ResponseCodeNotZero{response.Message}
	}
	return response, nil
}

type LoadFriendsResponse struct {
	Items []struct {
		DisplayName       string      `json:"displayName"`
		DisplayImageKey   int         `json:"displayImageKey"`
		LatestMessage     string      `json:"latestMessage"`
		LatestMessageTime time.Time   `json:"latestMessageTime"`
		UnReadAmount      int         `json:"unReadAmount"`
		ConversationID    int         `json:"conversationId"`
		Discriminator     string      `json:"discriminator"`
		UserID            interface{} `json:"userId"`
		AesKey            string      `json:"aesKey"`
		Muted             bool        `json:"muted"`
	} `json:"items"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (s *FriendshipService) LoadFriends() (*LoadFriendsResponse, error) {
	resp, err := s.client.client.Get(KahlaServer + "/friendship/MyFriends?orderByName=false")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, &ResponseStatusCodeNot200{Response: resp, StatusCode: resp.StatusCode}
	}
	j := json.NewDecoder(resp.Body)
	response := &LoadFriendsResponse{Code: -1}
	err = j.Decode(response)
	if err != nil {
		return response, &ResponseJsonDecodeError{response.Message, err}
	}
	if response.Code != 0 {
		return response, &ResponseCodeNotZero{response.Message}
	}
	return response, nil
}
