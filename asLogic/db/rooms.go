package db

import (
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/room"
)

// SaveRoom saves a room.Room{} struct to the Database
func SaveRoom(Room *room.Room) error {
	db := Open()
	tx, err := db.Begin()
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
	twitchChannel := Room.Alias

	_, err = stmt.Exec(alias, RoomID, twitchChannel)
	if err != nil {
		return err
	}
	tx.Commit()
	return nil
}

// GetRooms returns all saved Rooms from the DB mapped by the alias
func GetRooms() (rooms map[string]*room.Room, err error) {
	rooms = make(map[string]*room.Room)
	db := Open()
	rows, err := db.Query("SELECT room_alias, room_id, twitch_channel FROM rooms")
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

func GetTwitchRooms() (rooms map[string]string, err error) {
	nrooms, err := GetRooms()
	if err != nil {
		return nil, err
	}

	rooms = make(map[string]string)

	for _, v := range nrooms {
		rooms[v.TwitchChannel] = v.ID
	}

	return rooms, nil
}
