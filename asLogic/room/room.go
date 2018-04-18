package room

import "github.com/Nordgedanken/matrix-twitch-bridge/asLogic/twitch/websocket"

// Room contains the required Information about a Room
type Room struct {
	Alias         string
	ID            string
	TwitchChannel string
	TwitchWS      websocket.WebsocketHolder
}
