package websocket

import (
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	gws "github.com/gorilla/websocket"
	"kcaitech.com/kcserver/utils/str"
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
	wsWriteLock sync.Mutex
	wsReadLock  sync.Mutex
	isClose     bool
	handleClose func(int, string)
}

func newWs(ws *gws.Conn) *Ws {
	return &Ws{
		ws: ws,
	}
}

func Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*Ws, error) {
	upgrader := gws.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		EnableCompression: false,
	}
	conn, err := upgrader.Upgrade(w, r, responseHeader)
	if err != nil {
		return nil, err
	}
	return newWs(conn), nil
}

func NewClient(url string, requestHeader http.Header) (*Ws, error) {
	conn, _, err := gws.DefaultDialer.Dial(url, requestHeader)
	if err != nil {
		return nil, err
	}
	return newWs(conn), nil
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

func (ws *Ws) Lock() {
	ws.wsWriteLock.Lock()
	ws.wsReadLock.Lock()
}

func (ws *Ws) unlockWriteLock() {
	defer func() {
		recover()
	}()
	ws.wsWriteLock.Unlock()
}

func (ws *Ws) unlockReadLock() {
	defer func() {
		recover()
	}()
	ws.wsReadLock.Unlock()
}

func (ws *Ws) Unlock() {
	ws.unlockWriteLock()
	ws.unlockReadLock()
}

func isGwsClosedError(err error) bool {
	_, ok := err.(*gws.CloseError)
	return ok
}

const UseOfClosedNetworkConnectionWarn = "use of closed network connection"

func isClosedError(err error) bool {
	return isGwsClosedError(err) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || strings.Contains(err.Error(), UseOfClosedNetworkConnectionWarn)
}

func (ws *Ws) WriteMessageLock(needLock bool, messageType MessageType, data []byte) error {
	if needLock {
		ws.wsWriteLock.Lock()
		defer ws.wsWriteLock.Unlock()
	}
	if ws.isClose || ws.ws == nil {
		return ErrClosed
	}
	err := ws.ws.WriteMessage(int(messageType), data)
	if err != nil && isClosedError(err) {
		log.Println("ws-wr-msg", err)
		return ErrClosed
	}
	return err
}

func (ws *Ws) WriteJSONLock(needLock bool, v any) error {
	if needLock {
		ws.wsWriteLock.Lock()
		defer ws.wsWriteLock.Unlock()
	}
	if ws.isClose || ws.ws == nil {
		return ErrClosed
	}
	err := ws.ws.WriteJSON(v)
	if err != nil && isClosedError(err) {
		log.Println("ws-wr-json", err)
		return ErrClosed
	}
	return err
}

func (ws *Ws) ReadMessageLock(needLock bool) (MessageType, []byte, error) {
	if needLock {
		ws.wsReadLock.Lock()
		defer ws.wsReadLock.Unlock()
	}
	if ws.ws == nil {
		return MessageTypeNone, nil, ErrClosed
	}
	messageType, data, err := ws.ws.ReadMessage()
	if err != nil {
		log.Println("ws-rd-msg", err)
		if isClosedError(err) {
			return MessageTypeNone, data, ErrClosed
		}
		return MessageTypeNone, data, err
	}
	if messageType == int(MessageTypeText) || messageType == int(MessageTypeBinary) {
		return MessageType(messageType), data, nil
	}
	if ws.isClose {
		return MessageTypeNone, data, ErrClosed
	}
	return MessageTypeNone, data, errors.New("不支持的消息类型：" + str.IntToString(int64(messageType)))
}

func (ws *Ws) ReadJSONLock(needLock bool, v any) error {
	if needLock {
		ws.wsReadLock.Lock()
		defer ws.wsReadLock.Unlock()
	}
	if ws.isClose || ws.ws == nil {
		return ErrClosed
	}
	err := ws.ws.ReadJSON(v)
	if err != nil && isClosedError(err) {
		log.Println("ws-rd-json", err)
		return ErrClosed
	}
	return err
}

func (ws *Ws) WriteMessage(messageType MessageType, data []byte) error {
	return ws.WriteMessageLock(true, messageType, data)
}

func (ws *Ws) WriteJSON(v any) error {
	return ws.WriteJSONLock(true, v)
}

func (ws *Ws) ReadMessage() (MessageType, []byte, error) {
	return ws.ReadMessageLock(true)
}

func (ws *Ws) ReadJSON(v any) error {
	return ws.ReadJSONLock(true, v)
}

func (ws *Ws) Close() {
	log.Println("close ws")
	if ws.isClose || ws.ws == nil {
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
