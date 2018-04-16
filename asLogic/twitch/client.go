package twitch

import (
	"fmt"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/user"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	"github.com/gorilla/websocket"
	"os"
	"os/signal"
	"strings"
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

func Join(WS *websocket.Conn, channel string) error {
	// Join Room
	err := WS.WriteMessage(websocket.TextMessage, []byte("JOIN #"+channel))
	return err
}

func Send(WS *websocket.Conn, channel, message string) error {
	// Send Message
	util.Config.Log.Debugln(channel)
	err := WS.WriteMessage(websocket.TextMessage, []byte("PRIVMSG #"+channel+" :"+message))
	return err
}

// Listen answers to the PING messages by Twitch and relays messages to Matrix
func Listen(users map[string]*user.ASUser, rooms map[string]string) {
	go func() {
		defer close(util.Done)
		for {
			_, message, err := util.BotUser.TwitchWS.ReadMessage()
			if err != nil {
				util.Config.Log.Errorln(err)
				return
			}
			parsedMessage := parseMessage(fmt.Sprintf("%s", message))
			if parsedMessage != nil {
				switch parsedMessage.Command {
				case "PRIVMSG":
					room := rooms[parsedMessage.Channel]
					users[parsedMessage.Username].MXClient.SendText(room, parsedMessage.Message)
				case "PING":
					util.BotUser.TwitchWS.WriteControl(websocket.PongMessage, []byte(""), time.Now().Add(10*time.Second))
				default:
					util.Config.Log.Debugf("[TWITCH]: %+v\n", parsedMessage)
				}
			}

			util.Config.Log.Debugf("recv: %s\n", message)
		}
	}()
}

func parseMessage(message string) (parsedMessage *util.TMessage) {
	/*
		Actual Message from the Websocket:
		@badges=broadcaster/1;color=#D2691E;display-name=MTRNord;emotes=;id=3e969619-5312-4999-ba21-6d0ab81af8f5;mod=0;room-id=36031510;subscriber=0;tmi-sent-ts=1523458219318;turbo=0;user-id=36031510;user-type= :mtrnord!mtrnord@mtrnord.tmi.twitch.tv PRIVMSG #mtrnord :test
	*/

	parsedMessage = &util.TMessage{}
	if strings.HasPrefix(message, "@") {
		messageSplit := strings.Split(message, " ")
		parsedMessage.Tags = messageSplit[0]
		parsedMessage.Username = strings.TrimRight(strings.TrimLeft(messageSplit[1], ":"), "!")
		parsedMessage.Command = messageSplit[2]
		parsedMessage.Channel = messageSplit[3]
		parsedMessage.Message = strings.TrimLeft(messageSplit[4], ":")
	} else if strings.HasPrefix(message, "PING") {
		parsedMessage.Command = "PING"
		parsedMessage.Message = strings.Split(message, ":")[1]
	} else {
		parsedMessage = nil
	}

	return
}
