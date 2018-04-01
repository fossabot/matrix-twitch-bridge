package asLogic

import (
	"maunium.net/go/mautrix-appservice-go"
	"log"
	"github.com/fatih/color"
)

var config *appservice.Config

func Init()  {
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
	users map[string]string
	aliases map[string]string
}

func(q QueryHandler) QueryAlias(alias string) bool {
	// TODO implement logic
	return false
}

func(q QueryHandler) QueryUser(userID string) bool {
	// TODO implement logic
	return false
}