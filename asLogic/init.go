package asLogic

import (
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/db"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/room"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/twitch"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/user"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	"github.com/fatih/color"
	"github.com/matrix-org/gomatrix"
	"log"
	"maunium.net/go/mautrix-appservice-go"
	"regexp"
	"strings"
)

func Init() {
	var boldGreen = color.New(color.FgGreen).Add(color.Bold)
	appservice.GenerateRegistration("twitch", "twitch", true, true)
	boldGreen.Println("Please restart the Appservice with \"--config\" and \"--client_id\"-flag applied")
}

var realUsers map[string]*user.RealUser

func Run() error {
	var err error
	util.Config, err = appservice.Load(util.CfgFile)
	if err != nil {
		return err
	}

	queryHandler := QueryHandler{}
	queryHandler.aliases, err = db.GetRooms()
	if err != nil {
		return err
	}

	// TODO Make sure to load them from a DB!!!! (Even I forgot why I need this map...)
	queryHandler.twitchRooms = make(map[string]string)

	queryHandler.twitchUsers, err = db.GetTwitchUsers()
	if err != nil {
		return err
	}

	queryHandler.users, err = db.GetASUsers()
	if err != nil {
		return err
	}

	realUsers, err = db.GetRealUsers()
	if err != nil {
		return err
	}

	util.Config.Init(queryHandler)

	util.Config.Listen()

	//TODO INIT ROOM BRIDGES
	//TOKEN NEEDS TO BE A BOT
	//USERNAME NEEDS TO BE A BOT
	//twitch.Connect(token, username)
	//twitch.Listen(q.twitchUsers, q.twitchRooms)

	for {
		select {
		case event := <-util.Config.Events:
			switch event.Type {
			case "m.room.message":
				mxUser := realUsers[event.SenderID]
				if mxUser == nil {
					mxUser = &user.RealUser{}
					mxUser.Mxid = event.SenderID
					db.SaveRealUser(mxUser)
					// TODO Implement Auth logic and Queue the message for later!
					continue
				}
				if mxUser.TwitchWS == nil {
					if mxUser.TwitchToken != "" && mxUser.TwitchName != "" {
						mxUser.TwitchWS, err = twitch.Connect(mxUser.TwitchToken, mxUser.TwitchName)
						if err != nil {
							log.Println("[ERROR]: ", err)
							continue
						}
					}
				}

			}
		}
	}
	return nil
}

type QueryHandler struct {
	users       map[string]*user.ASUser
	aliases     map[string]*room.Room
	twitchUsers map[string]*user.ASUser
	twitchRooms map[string]string
}

func (q QueryHandler) QueryAlias(alias string) bool {
	if q.aliases[alias] != nil {
		return true
	}
	var tUsername string
	client, err := gomatrix.NewClient(util.Config.HomeserverURL, "", util.Config.Registration.AppToken)
	if err != nil {
		util.Config.Log.Errorln(err)
		return false
	}

	roomalias := strings.Split(strings.TrimLeft(alias, "#"), ":")[0]

	createRoomReq := &gomatrix.ReqCreateRoom{}
	// TODO Get actual Name
	for _, v := range util.Config.Registration.Namespaces.RoomAliases {
		r, err := regexp.Compile(v.Regex)
		if err != nil {
			util.Config.Log.Errorln(err)
			return false
		}
		if r.MatchString(alias) {
			tUsername := r.FindStringSubmatch(alias)[0]
			userdata, err := twitch.RequestUserData(tUsername)
			if err != nil {
				util.Config.Log.Errorln(err)
				return false
			}
			if userdata.Total == 0 {
				util.Config.Log.Errorln("user missing")
				return false
			}
			createRoomReq.Name = userdata.Users[0].DisplayName
			resp, err := client.UploadLink(userdata.Users[0].Logo)
			content := make(map[string]interface{})
			content["url"] = resp.ContentURI
			m_room_avatar_event := gomatrix.Event{
				Type:    "m.room.avatar",
				Content: content,
			}
			createRoomReq.InitialState = append(createRoomReq.InitialState, m_room_avatar_event)
			break
		}
	}
	createRoomReq.RoomAliasName = roomalias
	createRoomReq.Preset = "public_chat"
	resp, err := client.CreateRoom(createRoomReq)
	if err != nil {
		util.Config.Log.Errorln(err)
		return false
	}

	troom := &room.Room{
		Alias:         alias,
		ID:            resp.RoomID,
		TwitchChannel: tUsername,
	}
	q.aliases[alias] = troom
	err = db.SaveRoom(troom)
	if err != nil {
		util.Config.Log.Errorln(err)
		return false
	}
	return false
}

type registerAuth struct {
	Type string `json:"type"`
}

func (q QueryHandler) QueryUser(userID string) bool {
	if q.users[userID] != nil {
		return true
	}
	var tUsername string
	for _, v := range util.Config.Registration.Namespaces.UserIDs {
		r, err := regexp.Compile(v.Regex)
		if err != nil {
			util.Config.Log.Errorln(err)
			return false
		}
		if r.MatchString(userID) {
			tUsername = r.FindStringSubmatch(userID)[0]
			break
		}
	}

	check, err := twitch.CheckTwitchUser(tUsername)
	if err != nil {
		util.Config.Log.Errorln(err)
		return false
	}
	if !check {
		return false
	}
	asUser := user.ASUser{}
	asUser.Mxid = userID
	client, err := gomatrix.NewClient(util.Config.HomeserverURL, userID, util.Config.Registration.AppToken)
	if err != nil {
		util.Config.Log.Errorln(err)
		return false
	}
	asUser.MXClient = client
	MXusername := strings.Split(strings.TrimPrefix(userID, "@"), ":")[0]

	registerReq := gomatrix.ReqRegister{
		Username: MXusername,
		Auth: registerAuth{
			Type: "m.login.application_service",
		},
	}
	register, inter, err := client.Register(&registerReq)
	if err != nil {
		util.Config.Log.Errorln(err)
		return false
	}
	if inter != nil || register == nil {
		util.Config.Log.Errorln("Error encountered during user registration")
		return false
	}
	client.AppServiceUserID = userID
	userdata, err := twitch.RequestUserData(tUsername)
	if err != nil {
		util.Config.Log.Errorln(err)
		return false
	}
	if userdata.Total == 0 {
		util.Config.Log.Errorln("user missing")
		return false
	}
	client.SetDisplayName(userdata.Users[0].DisplayName)
	resp, err := client.UploadLink(userdata.Users[0].Logo)
	client.SetAvatarURL(resp.ContentURI)

	q.users[userID] = &asUser
	err = db.SaveASUser(q.users[userID])
	if err != nil {
		util.Config.Log.Errorln(err)
		return false
	}
	return true
}
