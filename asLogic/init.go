package asLogic

import (
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/room"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/twitch"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/user"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	"github.com/fatih/color"
	"github.com/matrix-org/gomatrix"
	"log"
	"maunium.net/go/mautrix-appservice-go"
	"strings"
)

func Init() {
	var boldGreen = color.New(color.FgGreen).Add(color.Bold)
	appservice.GenerateRegistration("twitch", "twitch", true, true)
	boldGreen.Println("Please restart the Appservice with \"--config\"-flag applied")
}

func Run(cfgFile string) error {
	var err error
	util.Config, err = appservice.Load(cfgFile)
	if err != nil {
		return err
	}

	queryHandler := QueryHandler{}
	queryHandler.twitchRooms = make(map[string]string)
	queryHandler.twitchUsers = make(map[string]*user.User)

	util.Config.Init(queryHandler)

	util.Config.Listen()
	twitch.Listen(queryHandler.twitchUsers, queryHandler.twitchRooms)

	for {
		select {
		case event := <-util.Config.Events:
			log.Println(event)
		}
	}
	return nil
}

type QueryHandler struct {
	users       map[string]*user.User
	aliases     map[string]*room.Room
	twitchUsers map[string]*user.User
	twitchRooms map[string]string
}

func (q QueryHandler) QueryAlias(alias string) bool {
	if q.aliases[alias] != nil {
		return true
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
	asUser := user.User{}
	asUser.Mxid = userID
	client, err := gomatrix.NewClient(util.Config.HomeserverURL, userID, util.Config.Registration.AppToken)
	if err != nil {
		util.Config.Log.Errorln(err)
		return false
	}
	asUser.MXClient = client
	username := strings.Split(strings.TrimPrefix(userID, "@"), ":")[0]

	registerReq := gomatrix.ReqRegister{
		Username: username,
		Auth: registerAuth{
			Type: "m.login.application_service",
		},
	}
	register, inter, err := asUser.MXClient.Register(&registerReq)
	if err != nil {
		util.Config.Log.Errorln(err)
		return false
	}
	if inter != nil || register == nil {
		util.Config.Log.Errorln("Error encountered during user registration")
		return false
	}
	client.AppServiceUserID = userID

	q.users[userID] = &asUser
	// TODO Link username to user on twitch (do some magic check if the user exists by crawling the channel page?) https://api.twitch.tv/kraken/users?login=<username>  DOC: https://dev.twitch.tv/docs/v5/
	return true
}
