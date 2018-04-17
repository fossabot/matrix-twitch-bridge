package twitch

import (
	"fmt"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/db"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/matrix_helper"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/queryHandler"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/twitch/api"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/user"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	"github.com/gorilla/websocket"
	"github.com/matrix-org/gomatrix"
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
					//Create AS User if needed and invite to room
					if queryHandler.QueryHandler().TwitchUsers[parsedMessage.Username] == nil {
						check, err := api.CheckTwitchUser(parsedMessage.Username)
						if err != nil {
							util.Config.Log.Errorln(err)
							return
						}
						if !check {
							return
						}

						for _, v := range util.Config.Registration.Namespaces.RoomAliases {
							// name magic
							pre := strings.Split(v.Regex, ".+")[0]
							suff := strings.Split(v.Regex, ".+")[1]
							asUser := &user.ASUser{}
							asUser.Mxid = pre + parsedMessage.Username + suff
							MXusername := strings.Split(strings.TrimPrefix(asUser.Mxid, "@"), ":")[0]
							client, err := gomatrix.NewClient(util.Config.HomeserverURL, asUser.Mxid, util.Config.Registration.AppToken)
							if err != nil {
								util.Config.Log.Errorln(err)
								return
							}
							asUser.MXClient = client

							err = matrix_helper.CreateUser(client, MXusername)
							if err != nil {
								util.Config.Log.Errorln(err)
								return
							}

							client.AppServiceUserID = asUser.Mxid

							userdata, err := api.RequestUserData(parsedMessage.Username)
							if err != nil {
								util.Config.Log.Errorln(err)
								return
							}
							if userdata.Total == 0 {
								util.Config.Log.Errorln("user missing")
								return
							}
							client.SetDisplayName(userdata.Users[0].DisplayName)
							resp, err := client.UploadLink(userdata.Users[0].Logo)
							client.SetAvatarURL(resp.ContentURI)

							queryHandler.QueryHandler().TwitchUsers[parsedMessage.Username] = asUser
							err = db.SaveUser(queryHandler.QueryHandler().TwitchUsers[parsedMessage.Username])
							if err != nil {
								util.Config.Log.Errorln(err)
							}
							break
						}
					}

					// Check if user needs to join the room
					joinedResp, err := util.BotUser.MXClient.JoinedMembers(room)
					if err != nil {
						util.Config.Log.Errorln(err)
						return
					}
					if _, ok := joinedResp.Joined[queryHandler.QueryHandler().TwitchUsers[parsedMessage.Username].Mxid]; !ok {
						queryHandler.QueryHandler().TwitchUsers[parsedMessage.Username].MXClient.JoinRoom(room, "", nil)
					}

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
