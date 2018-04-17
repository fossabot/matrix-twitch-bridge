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

func Send(WS *websocket.Conn, channel, messageRaw string) error {
	// Send Message
	message := "PRIVMSG #" + channel + " :" + messageRaw + "\r\n"
	util.Config.Log.Debugln("Message: ", message)
	util.Config.Log.Debugf("WS: %+v\n", WS.UnderlyingConn())
	err := WS.WriteMessage(websocket.TextMessage, []byte(message))
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

			util.Config.Log.Debugf("recv: %s\n", message)
			parsedMessage := parseMessage(fmt.Sprintf("%s", message))
			if parsedMessage != nil {
				switch parsedMessage.Command {
				case "PRIVMSG":
					room := queryHandler.QueryHandler().TwitchRooms[strings.TrimPrefix(parsedMessage.Channel, "#")]
					asUser := queryHandler.QueryHandler().TwitchUsers[parsedMessage.Username]

					//Create AS User if needed and invite to room
					if asUser == nil {
						check, err := api.CheckTwitchUser(parsedMessage.Username)
						if err != nil {
							util.Config.Log.Errorln(err)
							return
						}
						if !check {
							return
						}

						for _, v := range util.Config.Registration.Namespaces.UserIDs {
							// name magic
							pre := strings.Split(v.Regex, ".+")[0]
							suff := strings.Split(v.Regex, ".+")[1]
							asUser = &user.ASUser{}
							asUser.Mxid = pre + parsedMessage.Username + suff
							util.Config.Log.Debugln(asUser.Mxid)
							MXusername := strings.Split(strings.TrimPrefix(asUser.Mxid, "@"), ":")[0]
							util.Config.Log.Debugln(MXusername)
							client, err := gomatrix.NewClient(util.Config.HomeserverURL, asUser.Mxid, util.Config.Registration.AppToken)
							if err != nil {
								util.Config.Log.Errorln(err)
								return
							}
							asUser.MXClient = client
							asUser.TwitchName = parsedMessage.Username

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
							err = client.SetDisplayName(userdata.Users[0].DisplayName)
							if err != nil {
								util.Config.Log.Errorln(err)
							}
							resp, err := client.UploadLink(userdata.Users[0].Logo)
							if err != nil {
								util.Config.Log.Errorln(err)
							}
							err = client.SetAvatarURL(resp.ContentURI)
							if err != nil {
								util.Config.Log.Errorln(err)
							}

							queryHandler.QueryHandler().TwitchUsers[parsedMessage.Username] = asUser
							queryHandler.QueryHandler().Users[asUser.Mxid] = asUser
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
					mxid := asUser.Mxid
					if _, ok := joinedResp.Joined[mxid]; !ok {
						asUser.MXClient.JoinRoom(room, "", nil)
					}

					asUser.MXClient.SendText(room, parsedMessage.Message)
				case "PING":
					util.Config.Log.Debugln("[TWITCH]: Respond to Ping")
					util.BotUser.TwitchWS.WriteControl(websocket.PongMessage, []byte("\r\n"), time.Now().Add(10*time.Second))
				default:
					util.Config.Log.Debugf("[TWITCH]: %+v\n", parsedMessage)
				}
			}
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
		parsedMessage.Tags = strings.TrimSpace(messageSplit[0])
		parsedMessage.Username = strings.Split(strings.TrimLeft(strings.TrimSpace(messageSplit[1]), ":"), "!")[0]
		parsedMessage.Command = strings.TrimSpace(messageSplit[2])
		parsedMessage.Channel = strings.TrimSpace(messageSplit[3])
		rawMessageText := strings.TrimPrefix(message, messageSplit[0]+" "+messageSplit[1]+" "+messageSplit[2]+" "+messageSplit[3]+" ")
		parsedMessage.Message = strings.TrimLeft(strings.TrimSpace(rawMessageText), ":")
	} else if strings.HasPrefix(message, "PING") {
		parsedMessage.Command = "PING"
		parsedMessage.Message = strings.Split(message, ":")[1]
	} else {
		parsedMessage = nil
	}

	return
}
