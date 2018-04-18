package queryHandler

import (
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/matrix_helper"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/room"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/twitch/api"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/twitch/websocket/implementation"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/user"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	"github.com/matrix-org/gomatrix"
	"regexp"
	"strings"
	"sync"
)

// QueryHandler implements the interface appservice.QueryHandler{}
type queryHandler struct {
	Users       map[string]*user.ASUser
	RealUsers   map[string]*user.RealUser
	TwitchUsers map[string]*user.ASUser
	Aliases     map[string]*room.Room
	TwitchRooms map[string]string
}

var queryHandlerVar *queryHandler
var queryHandlerOnce sync.Once

func QueryHandler() *queryHandler {
	queryHandlerOnce.Do(func() {
		queryHandlerVar = &queryHandler{}
	})
	return queryHandlerVar
}

// QueryAlias is the logic that creates if needed a AS managed matrix room
// and tells the Homeserver if that room alias is managed by the AS
func (q queryHandler) QueryAlias(alias string) bool {
	if q.Aliases[alias] != nil {
		return true
	}
	var tUsername string
	client, err := gomatrix.NewClient(util.Config.HomeserverURL, "", util.Config.Registration.AppToken)
	if err != nil {
		util.Config.Log.Errorln(err)
		return false
	}

	roomalias := strings.Split(strings.TrimLeft(alias, "#"), ":")[0]

	var displayname string
	var logoURL string
	for _, v := range util.Config.Registration.Namespaces.RoomAliases {
		// name magic
		pre := strings.Split(v.Regex, ".+")[0]
		suff := strings.Split(v.Regex, ".+")[1]
		tUsername = strings.TrimSuffix(strings.TrimPrefix(alias, pre), suff)
		userdata, err := api.RequestUserData(tUsername)
		if err != nil {
			util.Config.Log.Errorln(err)
			return false
		}
		if userdata.Total == 0 {
			util.Config.Log.Errorln("user missing")
			return false
		}
		displayname = userdata.Users[0].DisplayName
		logoURL = userdata.Users[0].Logo
		break
	}

	resp, err := matrix_helper.CreateRoom(client, displayname, logoURL, roomalias, "public_chat")
	if err != nil {
		util.Config.Log.Errorln(err)
		return false
	}

	// TODO PUBLISH TO ROOM DICT

	troom := &room.Room{
		Alias:         alias,
		ID:            resp.RoomID,
		TwitchChannel: tUsername,
	}
	q.Aliases[alias] = troom
	q.TwitchRooms[troom.TwitchChannel] = troom.ID
	err = util.DB.SaveRoom(troom)
	if err != nil {
		util.Config.Log.Errorln(err)
	}

	util.BotUser.Mux.Lock()
	q.Aliases[alias].TwitchWS = &implementation.WebsocketHolder{
		Done:        make(chan struct{}),
		TwitchRooms: q.TwitchRooms,
		TwitchUsers: q.TwitchUsers,
		RealUsers:   q.RealUsers,
		Users:       q.Users,
	}
	err = q.Aliases[alias].TwitchWS.Join(tUsername)
	util.BotUser.Mux.Unlock()
	if err != nil {
		util.Config.Log.Errorln(err)
	}
	return true
}

// QueryUser is the logic that creates if needed a AS managed user
// and tells the Homeserver if that userID is managed by the AS
func (q queryHandler) QueryUser(userID string) bool {
	if q.Users[userID] != nil {
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

	check, err := api.CheckTwitchUser(tUsername)
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

	err = matrix_helper.CreateUser(client, MXusername)
	if err != nil {
		util.Config.Log.Errorln(err)
		return false
	}

	client.AppServiceUserID = userID
	userdata, err := api.RequestUserData(tUsername)
	if err != nil {
		util.Config.Log.Errorln(err)
		return false
	}
	if userdata.Total == 0 {
		util.Config.Log.Errorln("user missing")
		return false
	}
	err = client.SetDisplayName(userdata.Users[0].DisplayName + " (Twitch)")
	if err != nil {
		util.Config.Log.Errorln(err)
	}
	resp, err := client.UploadLink(userdata.Users[0].Logo)
	err = client.SetAvatarURL(resp.ContentURI)
	if err != nil {
		util.Config.Log.Errorln(err)
	}

	q.Users[userID] = &asUser
	err = util.DB.SaveUser(q.Users[userID])
	if err != nil {
		util.Config.Log.Errorln(err)
		return false
	}
	return true
}
