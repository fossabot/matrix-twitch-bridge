package implementation

import (
	"database/sql"
	dbHelper "github.com/Nordgedanken/matrix-twitch-bridge/asLogic/db/helper"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/room"
)

type DB struct {
	db *sql.DB
}

// SaveRoom saves a room.Room{} struct to the Database
func (d *DB) SaveRoom(Room *room.Room) error {
	if d.db == nil {
		d.db = dbHelper.Open()
	}
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare("INSERT INTO rooms (room_alias, room_id, twitch_channel) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	alias := Room.Alias
	RoomID := Room.ID
	twitchChannel := Room.TwitchChannel

	_, err = stmt.Exec(alias, RoomID, twitchChannel)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

// GetRooms returns all saved Rooms from the DB mapped by the alias
func (d *DB) GetRooms() (rooms map[string]*room.Room, err error) {
	rooms = make(map[string]*room.Room)
	if d.db == nil {
		d.db = dbHelper.Open()
	}
	rows, err := d.db.Query("SELECT room_alias, room_id, twitch_channel FROM rooms")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var RoomAlias string
		var RoomID string
		var TwitchChannel string
		err = rows.Scan(&RoomAlias, &RoomID, &TwitchChannel)
		if err != nil {
			return nil, err
		}

		room := &room.Room{
			Alias:         RoomAlias,
			ID:            RoomID,
			TwitchChannel: TwitchChannel,
		}

		rooms[RoomAlias] = room
	}

	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return rooms, nil
}

func (d *DB) GetTwitchRooms() (rooms map[string]string, err error) {
	nrooms, err := d.GetRooms()
	if err != nil {
		return nil, err
	}

	rooms = make(map[string]string)

	for _, v := range nrooms {
		rooms[v.TwitchChannel] = v.ID
	}

	return rooms, nil
}
