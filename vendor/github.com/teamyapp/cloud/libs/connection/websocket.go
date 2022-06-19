package connection

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var WebSocketUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocket struct {
	receiveMessageCh chan []byte
	sendMessageChan  chan []byte
	errorCh          chan error
	disconnectCh     chan bool
	conn             *websocket.Conn
}

var _ Connection = (*WebSocket)(nil)

func (w WebSocket) OnErrors() <-chan error {
	return w.errorCh
}

func (w WebSocket) OnMessageReceived() <-chan []byte {
	return w.receiveMessageCh
}

func (w WebSocket) SendMessage(message []byte) {
	w.sendMessageChan <- message
}

func (w WebSocket) OnClientDisconnect() <-chan bool {
	return w.disconnectCh
}

func (w WebSocket) Close() error {
	return w.conn.Close()
}

func NewWebSocket(conn *websocket.Conn) WebSocket {
	receiveMessageCh := make(chan []byte)
	sendMessageCh := make(chan []byte, 500)
	errorCh := make(chan error)
	disconnectCh := make(chan bool)
	conn.SetCloseHandler(func(code int, text string) error {
		disconnectCh <- true
		return nil
	})
	go func() {
		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				select {
				case errorCh <- err:
				default:
				}
				return
			}

			if mt != websocket.TextMessage && mt != websocket.BinaryMessage {
				continue
			}

			select {
			case receiveMessageCh <- message:
			default:
			}
		}
	}()
	go func() {
		for message := range sendMessageCh {
			err := conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Println(err)
				select {
				case errorCh <- err:
				default:
				}
				return
			}
		}
	}()
	return WebSocket{
		conn:             conn,
		receiveMessageCh: receiveMessageCh,
		sendMessageChan:  sendMessageCh,
		errorCh:          errorCh,
		disconnectCh:     disconnectCh,
	}
}
