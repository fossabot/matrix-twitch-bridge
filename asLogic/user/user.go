package user

import (
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/twitch/websocket"
	"github.com/matrix-org/gomatrix"
	"golang.org/x/oauth2"
	"net/http"
	"sync"
)

// ASUser contains the required Information for a User managed by the Appservice and holds a Matrix Client
type ASUser struct {
	Mxid       string
	TwitchName string
	MXClient   *gomatrix.Client
}

// RealUser contains the required Information for a Real User
// and holds a Matrix Client as well as a Websocket for twitch that allows to send messages to twitch
type RealUser struct {
	Mxid              string
	TwitchName        string
	TwitchTokenStruct *oauth2.Token
	TwitchHTTPClient  *http.Client
	TwitchWS          websocket.WebsocketHolder
	// Room holds a ID of a room with the Real User and the Bot
	Room string
	Mux  sync.Mutex
}

// BotUser contains the required Information for a Bot User
// and holds a Matrix Client as well as a Websocket for twitch that listens to the Channels
type BotUser struct {
	Mxid        string
	TwitchName  string
	TwitchToken string
	TwitchWS    websocket.WebsocketHolder
	Mux         sync.Mutex
	MXClient    *gomatrix.Client
}
