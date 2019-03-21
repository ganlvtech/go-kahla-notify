package kahla

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
)

const (
	EventTypeNewMessage = iota
	EventTypeNewFriendRequest
	EventTypeWereDeletedEvent
	EventTypeFriendAcceptedEvent
)

const (
	WebSocketStateNew = iota
	WebSocketStateConnected
	WebSocketStateDisconnected
	WebSocketStateClosed
)

type WebSocket struct {
	conn         *websocket.Conn
	State        int
	Event        chan interface{}
	Err          chan error
	Disconnected chan error
}

type KahlaEvent struct {
	Type int `json:"type"`
	Data interface{}
}

type InvalidEventTypeError struct {
	EventType int
}

func (i *InvalidEventTypeError) Error() string {
	return fmt.Sprintf("invalid event type: %d", i.EventType)
}

type NewMessageEvent struct {
	KahlaEvent
	ConversationID int `json:"conversationId"`
	Sender         struct {
		MakeEmailPublic   bool        `json:"makeEmailPublic"`
		Email             string      `json:"email"`
		ID                string      `json:"id"`
		Bio               interface{} `json:"bio"`
		NickName          string      `json:"nickName"`
		Sex               interface{} `json:"sex"`
		HeadImgFileKey    int         `json:"headImgFileKey"`
		PreferedLanguage  string      `json:"preferedLanguage"`
		AccountCreateTime string      `json:"accountCreateTime"`
		EmailConfirmed    bool        `json:"emailConfirmed"`
	} `json:"sender"`
	Content  string `json:"content"`
	AesKey   string `json:"aesKey"`
	Muted    bool   `json:"muted"`
	SentByMe bool   `json:"sentByMe"`
}

type NewFriendRequestEvent struct {
	KahlaEvent
	RequesterId string `json:"requesterId"`
}

type WereDeletedEvent KahlaEvent

type FriendAcceptedEvent KahlaEvent

func NewWebSocket() (*WebSocket) {
	w := new(WebSocket)
	w.Disconnected = make(chan error)
	w.Err = make(chan error)
	w.Event = make(chan interface{})
	w.State = WebSocketStateNew
	return w
}

func (w *WebSocket) Connect(serverPath string) error {
	var err error
	w.conn, _, err = websocket.DefaultDialer.Dial(serverPath, nil)
	if err != nil {
		return err
	}

	go w.runReceiveMessage()

	w.State = WebSocketStateConnected
	return nil
}

func (w *WebSocket) runReceiveMessage() {
	for {
		_, message, err := w.conn.ReadMessage()
		if err != nil {
			w.State = WebSocketStateDisconnected
			select {
			case w.Err <- err:
			default:
			}
			w.Disconnected <- err
			return
		}
		event1 := new(KahlaEvent)
		err = json.Unmarshal(message, event1)
		if err != nil {
			select {
			case w.Err <- err:
			default:
			}
			continue
		}
		var event interface{}
		switch event1.Type {
		case EventTypeNewMessage:
			event = new(NewMessageEvent)
		case EventTypeNewFriendRequest:
			event = new(NewFriendRequestEvent)
		case EventTypeWereDeletedEvent:
			event = new(WereDeletedEvent)
		case EventTypeFriendAcceptedEvent:
			event = new(FriendAcceptedEvent)
		default:
			err = &InvalidEventTypeError{event1.Type}
			select {
			case w.Err <- err:
			default:
			}
			continue
		}
		err = json.Unmarshal(message, event)
		if err != nil {
			select {
			case w.Err <- err:
			default:
			}
			continue
		}
		select {
		case w.Event <- event:
		default:
		}
	}
}

func (w *WebSocket) Close() {
	if w.conn != nil {
		_ = w.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		_ = w.conn.Close()
	}
	w.State = WebSocketStateClosed
}
