package twitch

import (
	"fmt"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/user"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"
)

func Connect(oauthToken, username string) (err error) {
	// Make sure to catch the Interrupt Signal to close the WS gracefully
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	util.Done = make(chan struct{})

	dialer := websocket.DefaultDialer
	util.WS, _, err = dialer.Dial("wss://irc-ws.chat.twitch.tv:443/irc", nil)

	go func() {
		for {
			select {
			case <-util.Done:
				return
			case <-interrupt:
				// Cleanly close the connection by sending a close message and then
				// waiting (with timeout) for the server to close the connection.
				err = util.WS.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				select {
				case <-util.Done:
				case <-time.After(time.Second):
				}
				return
			}
		}
	}()

	// Request needed IRC Capabilities https://dev.twitch.tv/docs/irc/#twitch-specific-irc-capabilities
	util.WS.WriteMessage(websocket.TextMessage, []byte("CAP REQ :twitch.tv/membership twitch.tv/tags"))

	//Login
	util.WS.WriteMessage(websocket.TextMessage, []byte("PASS oauth:"+oauthToken))
	util.WS.WriteMessage(websocket.TextMessage, []byte("NICK "+username))

	return
}

func Listen(users map[string]*user.User, rooms map[string]string) {
	go func() {
		defer close(util.Done)
		for {
			_, message, err := util.WS.ReadMessage()
			if err != nil {
				log.Println("[Error]", err)
				return
			}
			parsedMessage := parseMessage(fmt.Sprintf("%s", message))
			if parsedMessage != nil {
				switch parsedMessage.Command {
				case "PRIVMSG":
					room := rooms[parsedMessage.Channel]
					users[parsedMessage.Username].MXClient.SendText(room, parsedMessage.Message)
				case "PING":
					util.WS.WriteControl(websocket.PongMessage, []byte(""), time.Now().Add(10*time.Second))
				}
			}

			log.Printf("recv: %s", message)
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
		parsedMessage.Command = "Ping"
		parsedMessage.Message = strings.Split(message, ":")[1]
	} else {
		parsedMessage = nil
	}

	return
}
