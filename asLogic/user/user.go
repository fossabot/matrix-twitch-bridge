package user

import (
	"github.com/gorilla/websocket"
	"github.com/matrix-org/gomatrix"
)

type ASUser struct {
	Mxid       string
	TwitchName string
	MXClient   *gomatrix.Client
}

type RealUser struct {
	Mxid        string
	TwitchName  string
	TwitchToken string
	TwitchWS    *websocket.Conn
}

type BotUser struct {
	Mxid        string
	TwitchName  string
	TwitchToken string
	TwitchWS    *websocket.Conn
	MXClient    *gomatrix.Client
}
