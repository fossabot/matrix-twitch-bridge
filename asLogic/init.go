package asLogic

import (
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/room"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/user"
	"github.com/fatih/color"
	"github.com/matrix-org/gomatrix"
	"log"
	"maunium.net/go/mautrix-appservice-go"
	"strings"
)

var config *appservice.Config

func Init() {
	var boldGreen = color.New(color.FgGreen).Add(color.Bold)
	appservice.GenerateRegistration("twitch", "twitch", true, true)
	boldGreen.Println("Please restart the Appservice with \"--config\"-flag applied")
}

func Run(cfgFile string) error {
	var err error
	config, err = appservice.Load(cfgFile)
	if err != nil {
		return err
	}

	queryHandler := QueryHandler{}

	config.Init(queryHandler)

	config.Listen()

	for {
		select {
		case event := <-config.Events:
			log.Println(event)
		}
	}
	return nil
}

type QueryHandler struct {
	users   map[string]*user.User
	aliases map[string]*room.Room
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
	user := user.User{}
	user.Mxid = userID
	client, err := gomatrix.NewClient(config.HomeserverURL, userID, config.Registration.AppToken)
	if err != nil {
		config.Log.Errorln(err)
		return false
	}
	user.MXClient = client
	username := strings.Split(strings.TrimPrefix(userID, "@"), ":")[0]

	registerReq := gomatrix.ReqRegister{
		Username: username,
		Auth: registerAuth{
			Type: "m.login.application_service",
		},
	}
	register, inter, err := user.MXClient.Register(&registerReq)
	if err != nil {
		config.Log.Errorln(err)
		return false
	}
	if inter != nil || register == nil {
		config.Log.Errorln("Error encountered during user registration")
		return false
	}
	client.AppServiceUserID = userID

	// TODO Link username to user on twitch (do some magic check if the user exists by crawling the channel page?) https://api.twitch.tv/kraken/users?login=<username>  DOC: https://dev.twitch.tv/docs/v5/
	return true
}
