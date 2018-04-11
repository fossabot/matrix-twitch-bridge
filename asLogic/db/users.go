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
	TwitchName := user.TwitchName

	_, err = stmt.Exec(mxid, TwitchName)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}
