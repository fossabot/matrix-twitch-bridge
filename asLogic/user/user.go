package user

import (
	"github.com/gorilla/websocket"
	"github.com/matrix-org/gomatrix"
	"golang.org/x/oauth2"
	"net/http"
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
	TwitchWS          *websocket.Conn
	// Room holds a ID of a room with the Real User and the Bot
	Room string
}

// BotUser contains the required Information for a Bot User
// and holds a Matrix Client as well as a Websocket for twitch that listens to the Channels
type BotUser struct {
	Mxid        string
	TwitchName  string
	TwitchToken string
	TwitchWS    *websocket.Conn
	MXClient    *gomatrix.Client
}
