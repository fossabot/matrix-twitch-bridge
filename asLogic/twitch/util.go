package twitch

import (
	"encoding/json"
	"fmt"
	"github.com/Nordgedanken/matrix-twitch-bridge/asLogic/util"
	"net/http"
	"time"
)

// RequestUserData returns a Struct of User Information from twitch like the topic or display name
func RequestUserData(username string) (*UserJSON, error) {
	u := &UserJSON{}
	var httpCLient = &http.Client{
		Timeout: time.Second * 10,
	}
	req, err := http.NewRequest("GET", "https://api.twitch.tv/kraken/users?login="+username, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.twitchtv.v5+json")
	req.Header.Set("Client-ID", util.ClientID)

	res, err := httpCLient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.Body == nil {
		err := "Missing body"
		return nil, fmt.Errorf("%s", err)
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(u)
	return u, err
}

// CheckTwitchUser makes a request to the TwitchAPI to check if a channel exists or not
func CheckTwitchUser(username string) (bool, error) {
	u, err := RequestUserData(username)
	if err != nil {
		return false, err
	}

	if u.Total == 0 {
		return false, nil
	}

	for _, v := range u.Users {
		if v.Name == username {
			return true, nil
		}
	}

	return false, nil
}
