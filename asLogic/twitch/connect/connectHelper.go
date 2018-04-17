//Used to prevent import cycle
package connect

import (
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	"github.com/gorilla/websocket"
	"os"
	"os/signal"
	"time"
)

// Connect opens a Websocket and requests the needed Capabilities and does the Login
func Connect(oauthToken, username string) (WS *websocket.Conn, err error) {
	// Make sure to catch the Interrupt Signal to close the WS gracefully
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	if util.Done == nil {
		util.Done = make(chan struct{})
	}

	dialer := websocket.DefaultDialer
	WS, _, err = dialer.Dial("wss://irc-ws.chat.twitch.tv:443/irc", nil)

	go func() {
		for {
			select {
			case <-util.Done:
				return
			case <-interrupt:
				// Cleanly close the connection by sending a close message and then
				// waiting (with timeout) for the server to close the connection.
				err = WS.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				select {
				case <-util.Done:
				case <-time.After(time.Second):
				}
				os.Exit(0)
				return
			}
		}
	}()
	if err != nil {
		return
	}

	// Request needed IRC Capabilities https://dev.twitch.tv/docs/irc/#twitch-specific-irc-capabilities
	sendErr := WS.WriteMessage(websocket.TextMessage, []byte("CAP REQ :twitch.tv/membership twitch.tv/tags"))
	if sendErr != nil {
		err = sendErr
		return
	}

	// Login
	sendErr = WS.WriteMessage(websocket.TextMessage, []byte("PASS oauth:"+oauthToken))
	if sendErr != nil {
		err = sendErr
		return
	}
	sendErr = WS.WriteMessage(websocket.TextMessage, []byte("NICK "+username))
	if sendErr != nil {
		err = sendErr
		return
	}

	return
}