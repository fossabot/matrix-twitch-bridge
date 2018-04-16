package twitch

import (
	"fmt"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/queryHandler"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	"github.com/gorilla/websocket"
	"strings"
	"time"
)

func Send(WS *websocket.Conn, channel, message string) error {
	// Send Message
	util.Config.Log.Debugln(channel)
	err := WS.WriteMessage(websocket.TextMessage, []byte("PRIVMSG #"+channel+" :"+message))
	return err
}

// Listen answers to the PING messages by Twitch and relays messages to Matrix
func Listen() {
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
					room := queryHandler.QueryHandler().TwitchRooms[parsedMessage.Channel]
					queryHandler.QueryHandler().TwitchUsers[parsedMessage.Username].MXClient.SendText(room, parsedMessage.Message)
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
