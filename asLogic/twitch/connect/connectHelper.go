//Used to prevent import cycle
package connect

import (
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	"github.com/gorilla/websocket"
	"net"
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

	dialer := &websocket.Dialer{
		NetDial: func(network, addr string) (net.Conn, error) {
			netDialer := &net.Dialer{
				KeepAlive: time.Minute * 60,
			}
			return netDialer.Dial(network, addr)
		},
		HandshakeTimeout: 45 * time.Second,
	}
	WS, _, err = dialer.Dial("wss://irc-ws.chat.twitch.tv:443/irc", nil)

	go func() {
		for {
			select {
			case <-util.Done:
				util.Config.Log.Errorln("Done got closed")
				util.Config.Log.Errorln("Reconnecting WS")
				WS.Close()
				util.Done = make(chan struct{})
				Connect(oauthToken, username)
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
	sendErr := WS.WriteMessage(websocket.TextMessage, []byte("CAP REQ :twitch.tv/membership twitch.tv/tags\r\n"))
	if sendErr != nil {
		err = sendErr
		return
	}

	// Login
	sendErr = WS.WriteMessage(websocket.TextMessage, []byte("PASS oauth:"+oauthToken+"\r\n"))
	if sendErr != nil {
		err = sendErr
		return
	}
	sendErr = WS.WriteMessage(websocket.TextMessage, []byte("NICK "+username+"\r\n"))
	if sendErr != nil {
		err = sendErr
		return
	}

	WS.SetCloseHandler(func(code int, text string) error {
		WS.Close()
		Connect(oauthToken, username)
		return nil
	})

	return
}
