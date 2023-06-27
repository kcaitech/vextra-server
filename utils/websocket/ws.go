package websocket

import (
	"errors"
	gws "github.com/gorilla/websocket"
	"net/http"
)

var ErrClosed = errors.New("连接已关闭")

type MessageType int

const (
	MessageTypeNone   MessageType = 0
	MessageTypeText   MessageType = 1
	MessageTypeBinary MessageType = 2
)

type Ws struct {
	ws          *gws.Conn
	isClose     bool
	handleClose func(int, string)
}

func Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*Ws, error) {
	upgrader := gws.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, err
	}
	return &Ws{
		ws: conn,
	}, nil
}

func (ws *Ws) SetCloseHandler(handler func(code int, text string)) {
	ws.handleClose = handler
	ws.ws.SetCloseHandler(func(code int, text string) error {
		if ws.isClose {
			return nil
		}
		ws.isClose = true
		if handler != nil {
			handler(code, text)
		}
		return nil
	})
}

func (ws *Ws) WriteMessage(messageType MessageType, data []byte) error {
	if ws.isClose {
		return ErrClosed
	}
	return ws.ws.WriteMessage(int(messageType), data)
}

func (ws *Ws) WriteJSON(v any) error {
	if ws.isClose {
		return ErrClosed
	}
	return ws.ws.WriteJSON(v)
}

func (ws *Ws) ReadMessage() (MessageType, []byte, error) {
	messageType, data, err := ws.ws.ReadMessage()
	if err != nil {
		return MessageTypeNone, data, err
	}
	if messageType == int(MessageTypeText) || messageType == int(MessageTypeBinary) {
		return MessageType(messageType), data, nil
	}
	if ws.isClose {
		return MessageTypeNone, data, ErrClosed
	}
	return ws.ReadMessage()
}

func (ws *Ws) ReadJSON(v any) error {
	if ws.isClose {
		return ErrClosed
	}
	return ws.ws.ReadJSON(v)
}

func (ws *Ws) Close() {
	if ws.isClose {
		return
	}
	_ = ws.ws.Close()
	if ws.handleClose != nil {
		ws.handleClose(0, "")
	}
	ws.isClose = true
}

func (ws *Ws) IsClose() bool {
	return ws.isClose
}
