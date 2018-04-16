package db

import (
	"database/sql"
	"fmt"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

var db *sql.DB

// Init prepares the DBby opening it and creating the required tables if needed
func Init() (err error) {
	log.Println("Start setting up DB")
	var openErr error
	log.Println("DBFilePath: ", util.DbFile)
	db, openErr = sql.Open("sqlite3", util.DbFile)
	if openErr != nil {
		err = openErr
		return
	}

	log.Println("Creating DB Tables if needed")
	createTables := `CREATE TABLE IF NOT EXISTS users (id integer not null primary key, type text , mxid text, twitch_name text, twitch_token text, twitch_token_id text);
					CREATE TABLE IF NOT EXISTS tokens (id integer not null primary key, access_token text, token_type text, refresh_token text, expiry text);
					CREATE TABLE IF NOT EXISTS rooms (id integer not null primary key, room_alias text, room_id text, twitch_channel text);
					`
	_, execErr := db.Exec(createTables)
	if execErr != nil {
		err = fmt.Errorf("DB EXEC ERR: %s", execErr)
		return
	}
	log.Println("Finished setting DB Setup")
	return
}

// Open returns the in Init() created db variable
func Open() *sql.DB {
	if db != nil {
		return db
	}
	return nil
}
