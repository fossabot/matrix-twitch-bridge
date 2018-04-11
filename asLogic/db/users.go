package db

import (
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/user"
	"log"
)

func SaveASUser(user *user.ASUser) error {
	db := Open()
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO users (mxid, twitch_name) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	mxid := user.Mxid
	twitchName := user.TwitchName

	_, err = stmt.Exec(mxid, twitchName)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func GetASUsers() (map[string]*user.ASUser, error) {
	ASMap := make(map[string]*user.ASUser)

	db := Open()
	rows, err := db.Query("SELECT mxid, twitch_name FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var mxid string
		var twitchName string
		err = rows.Scan(&mxid, &twitchName)
		if err != nil {
			return nil, err
		}
		asUser := &user.ASUser{
			Mxid:       mxid,
			TwitchName: twitchName,
		}
		ASMap[mxid] = asUser
	}

	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return ASMap, nil
}
