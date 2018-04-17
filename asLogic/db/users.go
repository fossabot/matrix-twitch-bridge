package db

import (
	"database/sql"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/matrix_helper"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/twitch/connect"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/user"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	"github.com/gorilla/websocket"
	"github.com/matrix-org/gomatrix"
	"golang.org/x/oauth2"
	"strings"
	"time"
)

// SaveUser saves a User struct to the Database
func SaveUser(userA interface{}) error {
	util.Config.Log.Debugln("Opening DB")
	db := Open()

	util.Config.Log.Debugln("Beginning DB transaction")
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	util.Config.Log.Debugln("Prepare DB Statement")
	stmt, err := tx.Prepare("INSERT INTO users (type, mxid, twitch_name, twitch_token, twitch_token_id) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}

	defer stmt.Close()
	var mxid string
	var twitchName string
	var twitchToken string
	var twitch_token_id int64
	var Type string
	switch v := userA.(type) {
	case *user.ASUser:
		mxid = v.Mxid
		twitchName = v.TwitchName
		Type = "AS"
	case *user.RealUser:
		mxid = v.Mxid
		Type = "REAL"
		twitchName = v.TwitchName
		if v.TwitchTokenStruct != nil {
			expiry, err := v.TwitchTokenStruct.Expiry.MarshalText()
			if err != nil {
				return err
			}
			tokenResp, err := tx.Exec("INSERT INTO tokens (access_token, token_type, refresh_token, expiry) VALUES (?, ?, ?, ?)", v.TwitchTokenStruct.AccessToken, v.TwitchTokenStruct.Type(), v.TwitchTokenStruct.RefreshToken, string(expiry[:]))
			if err != nil {
				return err
			}
			twitch_token_id, err = tokenResp.LastInsertId()
			if err != nil {
				return err
			}
		}
	case *user.BotUser:
		Type = "BOT"
		mxid = v.Mxid
		twitchName = v.TwitchName
		twitchToken = v.TwitchToken
	}

	_, err = stmt.Exec(Type, mxid, twitchName, twitchToken, twitch_token_id)
	if err != nil {
		return err
	}

	util.Config.Log.Debugln("Commit to DB")
	err = tx.Commit()
	return err
}

type userTransportStruct struct {
	ASUsers   []*user.ASUser
	RealUsers []*user.RealUser
	BotUsers  []*user.BotUser
}

func getUsers() (users *userTransportStruct, err error) {
	transportStruct := &userTransportStruct{}
	db := Open()
	rows, err := db.Query("SELECT type, mxid, twitch_name, twitch_token, twitch_token_id FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var Type string
		var mxid string
		var twitchName string
		var twitchToken sql.NullString
		var twitchTokenID sql.NullString
		err = rows.Scan(&Type, &mxid, &twitchName, &twitchToken, &twitchTokenID)
		if err != nil {
			return nil, err
		}

		switch Type {
		case "AS":
			ASUser := &user.ASUser{
				Mxid:       mxid,
				TwitchName: twitchName,
			}
			transportStruct.ASUsers = append(transportStruct.ASUsers, ASUser)
		case "REAL":
			var TwitchToken *oauth2.Token
			if twitchTokenID.Valid {
				var accessToken string
				var tokenType string
				var refreshToken string
				var expiry sql.NullString
				expiryTime := time.Time{}
				db.QueryRow("SELECT access_token, token_type, refresh_token, expiry FROM tokens WHERE id = "+twitchTokenID.String, &accessToken, &tokenType, &refreshToken, &expiry)

				if expiry.Valid {
					err = expiryTime.UnmarshalText([]byte(expiry.String))
					if err != nil {
						return nil, err
					}
				}

				TwitchToken = &oauth2.Token{
					AccessToken:  accessToken,
					TokenType:    tokenType,
					RefreshToken: refreshToken,
				}
				if expiry.Valid {
					TwitchToken.Expiry = expiryTime
				}
			}
			var ws *websocket.Conn
			if TwitchToken != nil {
				ws, err = connect.Connect(TwitchToken.AccessToken, twitchName)
				if err != nil {
					return nil, err
				}
			}

			RealUser := &user.RealUser{
				Mxid:              mxid,
				TwitchTokenStruct: TwitchToken,
				TwitchName:        twitchName,
				TwitchWS:          ws,
			}

			transportStruct.RealUsers = append(transportStruct.RealUsers, RealUser)
		case "BOT":
			var TwitchToken string
			if twitchToken.Valid {
				TwitchToken = twitchToken.String
			}

			ws, err := connect.Connect(TwitchToken, twitchName)
			if err != nil {
				return nil, err
			}

			BotUser := &user.BotUser{
				Mxid:        mxid,
				TwitchToken: TwitchToken,
				TwitchName:  twitchName,
				TwitchWS:    ws,
			}
			transportStruct.BotUsers = append(transportStruct.BotUsers, BotUser)
		}
	}

	// get any error encountered during iteration
	err = rows.Err()
	return transportStruct, err
}

// GetASUsers returns all Users of type AS mapped by the MXID
func GetASUsers() (map[string]*user.ASUser, error) {
	ASMap := make(map[string]*user.ASUser)
	dbResp, err := getUsers()
	if err != nil {
		return nil, err
	}
	for _, v := range dbResp.ASUsers {
		if v.MXClient == nil {
			client, err := gomatrix.NewClient(util.Config.HomeserverURL, v.Mxid, util.Config.Registration.AppToken)
			if err != nil {
				return nil, err
			}

			v.MXClient = client
		}
		ASMap[v.Mxid] = v
	}

	return ASMap, nil
}

// GetTwitchUsers returns all Users of type AS mapped by the Twitch Channel Name
func GetTwitchUsers() (map[string]*user.ASUser, error) {
	TwitchMap := make(map[string]*user.ASUser)

	dbResp, err := getUsers()
	if err != nil {
		return nil, err
	}
	for _, v := range dbResp.ASUsers {
		if v.MXClient == nil {
			client, err := gomatrix.NewClient(util.Config.HomeserverURL, v.Mxid, util.Config.Registration.AppToken)
			if err != nil {
				return nil, err
			}

			v.MXClient = client
		}
		TwitchMap[v.TwitchName] = v
	}

	return TwitchMap, nil
}

// GetRealUsers returns all Users of type REAL mapped by the MXID
func GetRealUsers() (map[string]*user.RealUser, error) {
	RealMap := make(map[string]*user.RealUser)

	dbResp, err := getUsers()
	if err != nil {
		return nil, err
	}
	for _, v := range dbResp.RealUsers {
		RealMap[v.Mxid] = v
	}

	return RealMap, nil
}

// GetBotUser returns all Users of type BOT
func GetBotUser() (*user.BotUser, error) {
	dbResp, err := getUsers()
	if err != nil {
		return nil, err
	}
	if len(dbResp.BotUsers) >= 1 {
		bot := dbResp.BotUsers[0]

		client, err := gomatrix.NewClient(util.Config.HomeserverURL, bot.Mxid, util.Config.Registration.AppToken)
		if err != nil {
			return nil, err
		}

		bot.MXClient = client

		return bot, nil
	}

	var localpart = strings.TrimSuffix(strings.TrimPrefix(strings.Replace(util.Config.Registration.Namespaces.UserIDs[0].Regex, ".+", util.Config.Registration.SenderLocalpart, -1), "@"), ":"+util.Config.HomeserverDomain)
	var userID = strings.Replace(util.Config.Registration.Namespaces.UserIDs[0].Regex, ".+", util.Config.Registration.SenderLocalpart, -1)
	util.Config.Log.Debugln("Bot localpart: ", localpart)
	botUser := &user.BotUser{
		Mxid:        userID,
		TwitchName:  util.BotUName,
		TwitchToken: util.BotAToken,
	}

	util.Config.Log.Debugln("Creating gomatrix Client for the Bot User")
	client, err := gomatrix.NewClient(util.Config.HomeserverURL, userID, util.Config.Registration.AppToken)
	if err != nil {
		return nil, err
	}

	util.Config.Log.Debugln("Creating Bot User on the HomeServer")
	err = matrix_helper.CreateUser(client, localpart)
	if err != nil {
		return nil, err
	}

	util.Config.Log.Debugln("Adding Client to Bot Struct")
	botUser.MXClient = client

	util.Config.Log.Debugf("Saving Bot User to DB: %+v\n", botUser)
	SaveUser(botUser)

	return botUser, nil
}
