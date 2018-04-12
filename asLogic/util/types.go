package util

import (
	"github.com/gorilla/websocket"
	"maunium.net/go/mautrix-appservice-go"
)

var Config *appservice.Config
var WS *websocket.Conn
var Done chan struct{}
var CfgFile string
var DbFile string
var ClientID string

type TMessage struct {
	/*
		var parsedMessage = {
				message: null,
				tags: null,
				command: null,
				original: rawMessage,
				channel: null,
				username: null
			};
	*/
	Message  string
	Tags     string
	Command  string
	Original string
	Channel  string
	Username string
}
