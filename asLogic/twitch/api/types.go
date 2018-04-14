package api

import "time"

// UserJson defines the json for the `https://api.twitch.tv/kraken/users?login=<username>`
type UserJSON struct {
	Total int `json:"_total"`
	Users []struct {
		DisplayName string      `json:"display_name"`
		ID          string      `json:"_id"`
		Name        string      `json:"name"`
		Type        string      `json:"type"`
		Bio         interface{} `json:"bio"`
		CreatedAt   time.Time   `json:"created_at"`
		UpdatedAt   time.Time   `json:"updated_at"`
		Logo        string      `json:"logo"`
	} `json:"users"`
}
