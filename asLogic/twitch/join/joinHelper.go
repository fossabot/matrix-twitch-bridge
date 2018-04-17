//Used to prevent import cycle
package join

import (
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	"github.com/gorilla/websocket"
)

func Join(WS *websocket.Conn, channel string) error {
	// Join Room
	join := "JOIN #" + channel
	util.Config.Log.Debugln("Join Command: ", join)
	joinByte := []byte(join)
	util.Config.Log.Debugln("Join Command Bytes: ", join)
	err := WS.WriteMessage(websocket.TextMessage, joinByte)
	return err
}
