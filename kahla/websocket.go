package kahla

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"time"
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
	Event        chan interface{}
	State        int
	StateChanged chan int
}

type Event struct {
	Type int `json:"type"`
}

type InvalidEventTypeError struct {
	EventType int
}

func (i *InvalidEventTypeError) Error() string {
	return fmt.Sprintf("invalid event type: %d", i.EventType)
}

type NewMessageEvent struct {
	Event
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
	Event
	RequesterId string `json:"requesterId"`
}

type WereDeletedEvent Event

type FriendAcceptedEvent Event

func NewWebSocket() *WebSocket {
	w := new(WebSocket)
	w.Event = make(chan interface{}, 10)
	w.StateChanged = make(chan int)
	w.changeState(WebSocketStateNew)
	return w
}

func (w *WebSocket) changeState(state int) {
	w.State = state
	select {
	case w.StateChanged <- state:
	default:
	}
}

// https://github.com/gorilla/websocket/blob/master/examples/echo/client.go
// wss://stargate.aiursoft.com/Listen/Channel?Id=&Key=
// You should get message from w.Event.
// This is a synchronize call, it returns when connection closed.
func (w *WebSocket) Connect(serverPath string, interrupt chan struct{}) error {
	// Connect
	var err error
	w.conn, _, err = websocket.DefaultDialer.Dial(serverPath, nil)
	if err != nil {
		return err
	}
	// close connection when return
	defer w.conn.Close()
	w.changeState(WebSocketStateConnected)

	// Main message loop in another goroutine
	done := make(chan struct{})
	errChan := make(chan error)
	go w.runReceiveMessage(done, errChan)

	ticker := time.NewTicker(45 * time.Second)
	defer ticker.Stop()

	// wait connection close or interrupt
	for {
		select {
		case <-done:
			// connection closed
			w.changeState(WebSocketStateDisconnected)
			return nil
		case err := <-errChan:
			// error
			w.changeState(WebSocketStateDisconnected)
			return err
		case <-ticker.C:
			// heartbeat
			err := w.conn.WriteMessage(websocket.TextMessage, []byte{})
			if err != nil {
				w.changeState(WebSocketStateDisconnected)
				return err
			}
		case <-interrupt:
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := w.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				// We must set state to closed instead of disconnected, even write close message failed.
				w.changeState(WebSocketStateClosed)
				return err
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			w.changeState(WebSocketStateClosed)
			return nil
		}
	}
}

func (w *WebSocket) runReceiveMessage(done chan<- struct{}, errChan chan<- error) {
	// done when main loop exit
	defer close(done)
	defer close(errChan)
	for {
		_, message, err := w.conn.ReadMessage()
		if err != nil {
			// send error and exit
			errChan <- err
			return
		}
		event, err := DecodeWebSocketEvent(message)
		if err != nil {
			errChan <- err
			return
		}
		w.Event <- event
	}
}

func DecodeWebSocketEvent(message []byte) (interface{}, error) {
	var err error
	event1 := &Event{}
	err = json.Unmarshal(message, event1)
	if err != nil {
		return event1, err
	}
	var event interface{}
	switch event1.Type {
	case EventTypeNewMessage:
		event = &NewMessageEvent{}
	case EventTypeNewFriendRequest:
		event = &NewFriendRequestEvent{}
	case EventTypeWereDeletedEvent:
		event = &WereDeletedEvent{}
	case EventTypeFriendAcceptedEvent:
		event = &FriendAcceptedEvent{}
	default:
		return event1, &InvalidEventTypeError{event1.Type}
	}
	err = json.Unmarshal(message, event)
	if err != nil {
		return event, err
	}
	return event, nil
}
