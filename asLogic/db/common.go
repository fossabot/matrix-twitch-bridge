package db

import (
	"database/sql"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	"sync"
)

var dbOnce sync.Once
var db *sql.DB

// Init prepares the DBby opening it and creating the required tables if needed
func Init() (err error) {
	dbOnce.Do(func() {
		db, err = sql.Open("sqlite3", util.DbFile)
		createTables := `CREATE TABLE IF NOT EXISTS users (id integer not null primary key, type text , mxid text, twitch_name text, twitch_token text);
						 CREATE TABLE IF NOT EXISTS rooms (id integer not null primary key, room_alias text, room_id text, twitch_channel text);
						 `
		_, err = db.Exec(createTables)
	})
	return
}

// Open returns the in Init() created db variable
func Open() *sql.DB {
	if db != nil {
		return db
	}
	return nil
}
