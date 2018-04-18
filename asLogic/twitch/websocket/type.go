package websocket

import "github.com/gorilla/websocket"

type WebsocketHolder interface {
	Send(channel, messageRaw string) error
	Join(channel string) error
	Connect(oauthToken, username string) (err error)
	GetWS() *websocket.Conn
	Listen()
}
