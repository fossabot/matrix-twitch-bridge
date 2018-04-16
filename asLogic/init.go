package asLogic

import (
	"fmt"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/db"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/queryHandler"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/twitch"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/twitch/login"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/user"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	"github.com/fatih/color"
	"github.com/gorilla/mux"
	"log"
	"maunium.net/go/maulogger"
	"maunium.net/go/mautrix-appservice-go"
	"net/http"
	"os"
)

// Init starts the interactive Config generator and exits
func Init() {
	var boldGreen = color.New(color.FgGreen).Add(color.Bold)
	appservice.GenerateRegistration("twitch", "appservice-twitch", true, true)
	boldGreen.Println("Please restart the Twitch-Appservice with \"--client_id\"-flag applied")
}

func prepareRun() error {
	var err error

	util.Config, err = appservice.Load(util.CfgFile)
	if err != nil {
		return err
	}

	util.Config.Registration, err = appservice.LoadRegistration(util.Config.RegistrationPath)

	util.Config.Log = maulogger.Create()
	util.Config.LogConfig.Debug = true
	util.Config.LogConfig.Configure(util.Config.Log)
	util.Config.Log.Debugln("Logger initialized successfully.")

	util.Config.Log.Debugln("Creating queryHandler.")
	qHandler := queryHandler.QueryHandler()

	util.Config.Log.Debugln("Loading Twitch Rooms from DB.")
	qHandler.TwitchRooms, err = db.GetTwitchRooms()
	if err != nil {
		return err
	}

	util.Config.Log.Debugln("Loading Rooms from DB.")
	qHandler.Aliases, err = db.GetRooms()
	if err != nil {
		return err
	}

	util.Config.Log.Debugln("Loading Twitch Users from DB.")
	qHandler.TwitchUsers, err = db.GetTwitchUsers()
	if err != nil {
		return err
	}

	util.Config.Log.Debugln("Loading AS Users from DB.")
	qHandler.Users, err = db.GetASUsers()
	if err != nil {
		return err
	}

	util.Config.Log.Debugln("Loading Real Users from DB.")
	qHandler.RealUsers, err = db.GetRealUsers()
	if err != nil {
		return err
	}

	util.Config.Log.Debugln("Loading Bot User from DB.")
	util.BotUser, err = db.GetBotUser()
	if err != nil {
		return err
	}

	util.Config.Log.Infoln("Init...")
	util.Config.Log.Close()
	_, err = util.Config.Init(qHandler)
	if err != nil {
		log.Fatalln(err)
	}
	util.Config.Log.Infoln("Init Done...")

	util.Config.Log.Infoln("Starting public server...")
	r := mux.NewRouter()
	r.HandleFunc("/callback", login.Callback).Methods(http.MethodGet)

	go func() {
		var err error
		if len(util.TLSCert) == 0 || len(util.TLSKey) == 0 {
			err = fmt.Errorf("You need to have a SSL Cert!")
		} else {
			err = http.ListenAndServeTLS(util.Publicaddress, util.TLSCert, util.TLSKey, r)
		}
		if err != nil {
			util.Config.Log.Fatalln("Error while listening:", err)
			os.Exit(1)
		}
	}()

	util.Config.Log.Infoln("Starting Appservice Server...")
	util.Config.Listen()

	return nil
}

// Run starts the actual Appservice to let it listen to both ends
func Run() error {
	err := prepareRun()
	if err != nil {
		return err
	}

	util.BotUser.TwitchWS, err = twitch.Connect(util.BotUser.TwitchToken, util.BotUser.TwitchName)
	if err != nil {
		return err
	}

	twitch.Listen(queryHandler.QueryHandler().TwitchUsers, queryHandler.QueryHandler().TwitchRooms)

	for {
		select {
		case event := <-util.Config.Events:
			switch event.Type {
			case "m.room.message":
				qHandler := queryHandler.QueryHandler()
				mxUser := qHandler.RealUsers[event.SenderID]
				if mxUser == nil {
					mxUser = &user.RealUser{}
					mxUser.Mxid = event.SenderID
					db.SaveUser(mxUser, "REAL")
					login.SendLoginURL(mxUser)
					continue
				}
				if mxUser.TwitchWS == nil {
					if mxUser.TwitchTokenStruct.AccessToken != "" && mxUser.TwitchName != "" {
						mxUser.TwitchWS, err = twitch.Connect(mxUser.TwitchTokenStruct.AccessToken, mxUser.TwitchName)
						if err != nil {
							util.Config.Log.Errorln(err)
							continue
						}
					}
				}

			}
		}
	}
}
