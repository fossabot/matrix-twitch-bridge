package user

import "github.com/matrix-org/gomatrix"

type User struct {
	Mxid       string
	TwitchName string
	MXClient   *gomatrix.Client
}
