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

	return nil
}

// Run starts the actual Appservice to let it listen to both ends
func Run() error {
	err := prepareRun()
	if err != nil {
		return err
	}

	util.Config.Log.Debugln("Start Connecting BotUser to Twitch as: ", util.BotUser.TwitchName)
	util.BotUser.TwitchWS, err = twitch.Connect(util.BotUser.TwitchToken, util.BotUser.TwitchName)
	if err != nil {
		return err
	}

	util.Config.Log.Debugln("Start letting BotUser listen to Twitch")
	twitch.Listen(queryHandler.QueryHandler().TwitchUsers, queryHandler.QueryHandler().TwitchRooms)

	for v := range queryHandler.QueryHandler().TwitchRooms {
		err = twitch.Join(util.BotUser.TwitchWS, v)
		if err != nil {
			return err
		}
	}

	go func() {
		for {
			select {
			case event := <-util.Config.Events:
				util.Config.Log.Debugln("Got Event")
				switch event.Type {
				case "m.room.message":

					qHandler := queryHandler.QueryHandler()
					for _, v := range qHandler.Aliases {
						if v.ID == event.RoomID {
							if event.SenderID != util.BotUser.MXClient.UserID {
								err = useEvent(event)
								if err != nil {
									util.Config.Log.Errorln(err)
								}
							}
						}
					}
					continue

				}
			}
		}
	}()

	util.Config.Log.Infoln("Starting Appservice Server...")
	util.Config.Listen()

	select {}
}

func useEvent(event appservice.Event) error {
	qHandler := queryHandler.QueryHandler()
	mxUser := qHandler.RealUsers[event.SenderID]
	if mxUser == nil {
		qHandler.RealUsers[event.SenderID] = &user.RealUser{}
		mxUser := qHandler.RealUsers[event.SenderID]
		mxUser.Mxid = event.SenderID
		mxUser.TwitchTokenStruct = nil
		err := login.SendLoginURL(mxUser)
		if err != nil {
			return err
		}
		return nil
	} else if mxUser.TwitchTokenStruct == nil {
		err := login.SendLoginURL(mxUser)
		if err != nil {
			return err
		}
	}
	db.SaveUser(mxUser, "REAL")
	if mxUser.TwitchTokenStruct.AccessToken != "" && mxUser.TwitchName != "" {
		if mxUser.TwitchWS == nil {
			var err error
			mxUser.TwitchWS, err = twitch.Connect(mxUser.TwitchTokenStruct.AccessToken, mxUser.TwitchName)
			if err != nil {
				return err
			}
		} else {
			var err error
			mxUser.TwitchWS, err = twitch.Connect(mxUser.TwitchTokenStruct.AccessToken, mxUser.TwitchName)
			if err != nil {
				return err
			}
		}
		for _, v := range qHandler.Aliases {
			if v.ID == event.RoomID {
				err := twitch.Join(mxUser.TwitchWS, v.TwitchChannel)
				if err != nil {
					return err
				}
				if event.Content["msgtype"] == "m.text" {
					err = twitch.Send(mxUser.TwitchWS, v.TwitchChannel, event.Content["body"].(string))
					if err != nil {
						return err
					}
				} else {
					resp, err := util.BotUser.MXClient.GetDisplayName(event.SenderID)
					if err != nil {
						return err
					}
					_, err = util.BotUser.MXClient.SendNotice(event.RoomID, resp.DisplayName+": Please use Text only as Twitch doesn't support any other Media Format!")
					if err != nil {
						return err
					}
				}
			}
		}

	}
	return nil
}
