package asLogic

import (
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/room"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/user"
	"github.com/fatih/color"
	"log"
	"maunium.net/go/mautrix-appservice-go"
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

func (q QueryHandler) QueryUser(userID string) bool {
	if q.users[userID] != nil {
		return true
	}
	return false
}
