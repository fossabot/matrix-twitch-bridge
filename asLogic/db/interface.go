package db

import (
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/room"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/user"
)

type Handler interface {
	SaveRoom(Room *room.Room) error
	GetRooms() (rooms map[string]*room.Room, err error)
	GetTwitchRooms() (rooms map[string]string, err error)

	SaveUser(userA interface{}) error
	GetASUsers() (map[string]*user.ASUser, error)
	GetTwitchUsers() (map[string]*user.ASUser, error)
	GetRealUsers() (map[string]*user.RealUser, error)
	GetBotUser() (*user.BotUser, error)
}
