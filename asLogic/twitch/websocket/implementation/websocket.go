//Used to prevent import cycle
package implementation

import (
	"fmt"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/matrix_helper"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/queryHandler"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/twitch/api"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/user"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	"github.com/gorilla/websocket"
	"github.com/matrix-org/gomatrix"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"
)

type WebsocketHolder struct {
	// Websocket
	WS *websocket.Conn
	// Done is used to gracefully exit all WS connections
	Done chan struct{}
}

func (w *WebsocketHolder) Send(channel, messageRaw string) error {
	// Send Message
	message := "PRIVMSG #" + channel + " :" + messageRaw + "\r\n"
	deadline := time.Now().Add(time.Second * 5)
	err := w.WS.SetWriteDeadline(deadline)
	if err != nil {
		return err
	}
	err = w.WS.WriteMessage(websocket.TextMessage, []byte(message))
	return err
}

func (w *WebsocketHolder) Join(channel string) error {
	// Join Room
	w.WS.SetWriteDeadline(time.Now().Add(time.Minute * 2))
	join := "JOIN #" + channel + "\r\n"
	util.Config.Log.Debugln("Join Command: ", join)
	joinByte := []byte(join)
	util.Config.Log.Debugln("Join Command Bytes: ", join)
	err := w.WS.WriteMessage(websocket.TextMessage, joinByte)
	return err
}

// Connect opens a Websocket and requests the needed Capabilities and does the Login
func (w *WebsocketHolder) Connect(oauthToken, username string) (err error) {
	// Make sure to catch the Interrupt Signal to close the WS gracefully
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	dialer := &websocket.Dialer{
		NetDial: func(network, addr string) (net.Conn, error) {
			netDialer := &net.Dialer{
				KeepAlive: time.Minute * 60,
			}
			return netDialer.Dial(network, addr)
		},
		HandshakeTimeout: 45 * time.Second,
	}
	w.WS, _, err = dialer.Dial("wss://irc-ws.chat.twitch.tv:443/irc", nil)

	go func() {
		for {
			select {
			case <-w.Done:
				util.Config.Log.Errorln("Done got closed")
				util.Config.Log.Errorln("Reconnecting WS")
				err = w.WS.Close()
				w.WS = nil
				w.Done = make(chan struct{})
				err = w.Connect(oauthToken, username)
				return
			case <-interrupt:
				// Cleanly close the connection by sending a close message and then
				// waiting (with timeout) for the server to close the connection.
				err = w.WS.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				select {
				case <-w.Done:
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
	sendErr := w.WS.WriteMessage(websocket.TextMessage, []byte("CAP REQ :twitch.tv/membership twitch.tv/tags\r\n"))
	if sendErr != nil {
		err = sendErr
		return
	}

	// Login
	sendErr = w.WS.WriteMessage(websocket.TextMessage, []byte("PASS oauth:"+oauthToken+"\r\n"))
	if sendErr != nil {
		err = sendErr
		return
	}
	sendErr = w.WS.WriteMessage(websocket.TextMessage, []byte("NICK "+username+"\r\n"))
	if sendErr != nil {
		err = sendErr
		return
	}

	w.WS.SetCloseHandler(func(code int, text string) error {
		err := w.WS.Close()
		if err != nil {
			return err
		}
		err = w.Connect(oauthToken, username)
		return err
	})

	return
}

// Listen answers to the PING messages by Twitch and relays messages to Matrix
func (w *WebsocketHolder) Listen() {
	go func() {
		defer close(w.Done)
		for {
			_, message, err := util.BotUser.TwitchWS.GetWS().ReadMessage()
			if err != nil {
				util.Config.Log.Errorln(err)
				return
			}

			util.Config.Log.Debugf("recv: %s\n", message)
			parsedMessage := parseMessage(fmt.Sprintf("%s", message))
			if parsedMessage != nil {
				switch parsedMessage.Command {
				case "PRIVMSG":
					real := false
					for _, v := range queryHandler.QueryHandler().RealUsers {
						if parsedMessage.Username == v.TwitchName {
							real = true
							break
						}
					}
					if real {
						continue
					}
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
							err = client.SetDisplayName(userdata.Users[0].DisplayName + " (Twitch)")
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
							err = util.DB.SaveUser(queryHandler.QueryHandler().TwitchUsers[parsedMessage.Username])
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
					util.BotUser.Mux.Lock()
					util.BotUser.TwitchWS.GetWS().WriteControl(websocket.PongMessage, []byte("\r\n"), time.Now().Add(10*time.Second))
					util.BotUser.Mux.Unlock()
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

func (w *WebsocketHolder) GetWS() *websocket.Conn {
	return w.WS
}
