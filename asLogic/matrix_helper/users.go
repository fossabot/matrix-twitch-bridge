package matrix_helper

import (
	"fmt"
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
	if err != nil {
		return err
	}
	if inter != nil || register == nil {
		return fmt.Errorf("%s", "Error encountered during user registration")
	}
	return nil
}
