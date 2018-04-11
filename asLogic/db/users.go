package db

import (
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/user"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	"github.com/matrix-org/gomatrix"
)

func SaveASUser(user *user.ASUser) error {
	db := Open()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO users (type, mxid, twitch_name) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	mxid := user.Mxid
	twitchName := user.TwitchName

	_, err = stmt.Exec("AS", mxid, twitchName)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func SaveRealUser(user *user.RealUser) error {
	db := Open()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO users (type, mxid, twitch_name, twitch_token) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	mxid := user.Mxid
	twitchName := user.TwitchName
	twitchToken := user.TwitchToken

	_, err = stmt.Exec("REAL", mxid, twitchName, twitchToken)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func SaveBotUser(user *user.BotUser) error {
	db := Open()
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO users (type, mxid, twitch_name, twitch_token) VALUES (?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	mxid := user.Mxid
	twitchName := user.TwitchName
	twitchToken := user.TwitchToken

	_, err = stmt.Exec("BOT", mxid, twitchName, twitchToken)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

type UserTransportStruct struct {
	ASUsers   []*user.ASUser
	RealUsers []*user.RealUser
	BotUsers  []*user.BotUser
}

func getUsers() (users *UserTransportStruct, err error) {
	transportStruct := &UserTransportStruct{}
	db := Open()
	rows, err := db.Query("SELECT type, mxid, twitch_name, twitch_token FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var Type string
		var mxid string
		var twitchName string
		var twitchToken string
		err = rows.Scan(&Type, &mxid, &twitchName, &twitchToken)
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
		}
	}

	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return transportStruct, nil
}

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
