package db

import (
	"database/sql"
	"sync"
)

var dbOnce sync.Once
var db *sql.DB

func Init(dbPath string) (err error) {
	dbOnce.Do(func() {
		db, err = sql.Open("sqlite3", dbPath)
		createTables := `CREATE TABLE IF NOT EXISTS users (id integer not null primary key, type text , mxid text, twitch_name text, twitch_token text);
						 CREATE TABLE IF NOT EXISTS rooms (id integer not null primary key, room_alias text, room_id text, twitch_channel text);
						 `
		_, err = db.Exec(createTables)
	})
	return
}

func Open() *sql.DB {
	if db != nil {
		return db
	}
	return nil
}
