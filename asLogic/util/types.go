package util

import (
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/user"
	"maunium.net/go/mautrix-appservice-go"
)

// Config makes the appservice accessible everywhere in the Golang Code
var Config *appservice.Config

// BotUser exposes the Bot User to the complete go code
var BotUser *user.BotUser

var BotAToken string

var BotUName string

// Done is used to gracefully exit all WS connections
var Done chan struct{}

// CfgFile holds the location of the Config File
var CfgFile string

// DbFile holds the location of the Database file
var DbFile string

// ClientID holds the client_id of the Twitch App needed to use the API as well as generate Login URLs
var ClientID string

// ClientSecret holds the client_secret of the Twitch App needed to use the API as well as generate Login URLs
var ClientSecret string

var TLSCert string

var TLSKey string

var Publicaddress string

// TMessage is a struct with information about a Message send by Twitch
type TMessage struct {
	Message  string
	Tags     string
	Command  string
	Original string
	Channel  string
	Username string
}
