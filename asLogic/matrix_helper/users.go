package matrix_helper

import (
	"fmt"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	"github.com/matrix-org/gomatrix"
)

type registerAuth struct {
	Type string `json:"type"`
}

func CreateUser(client *gomatrix.Client, username string) error {
	registerReq := gomatrix.ReqRegister{
		Username: username,
		Auth: registerAuth{
			Type: "m.login.application_service",
		},
	}

	register, inter, err := client.Register(&registerReq)
	util.Config.Log.Debugln("Wrapped Matrix Err: ", err.(gomatrix.HTTPError).WrappedError.(gomatrix.RespError).Err)
	if err != nil && err.(gomatrix.HTTPError).WrappedError.(gomatrix.RespError).Err != "M_USER_IN_USE" {
		return err
	}
	if inter != nil || register == nil && err.(*gomatrix.HTTPError).WrappedError.(gomatrix.RespError).Err != "M_USER_IN_USE" {
		return fmt.Errorf("%s", "Error encountered during user registration")
	}
	return nil
}
