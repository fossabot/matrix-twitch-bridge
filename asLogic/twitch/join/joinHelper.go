//Used to prevent import cycle
package join

import "github.com/gorilla/websocket"

func Join(WS *websocket.Conn, channel string) error {
	// Part Room
	err := WS.WriteMessage(websocket.TextMessage, []byte("PART #"+channel))
	if err != nil {
		return err
	}
	// Join Room
	err = WS.WriteMessage(websocket.TextMessage, []byte("JOIN #"+channel))
	return err
}
