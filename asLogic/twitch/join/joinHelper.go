//Used to prevent import cycle
package join

import "github.com/gorilla/websocket"

func Join(WS *websocket.Conn, channel string) error {
	// Join Room
	err := WS.WriteMessage(websocket.TextMessage, []byte("JOIN #"+channel))
	return err
}
