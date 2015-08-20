package google

import (
	"encoding/json"
	"errors"
	"net/http"
)

const (
	googleVerifyURL = "https://www.googleapis.com/oauth2/v2/userinfo"
)

// User represents the data for a given Google User
type User struct {
	GoogleID      string `json:"id"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	GooglePlusURL string `json:"link"`
	AvatarURL     string `json:"picture"`
	Gender        string `json:"gender"`
	Locale        string `json:"locale"`
}

// VerifyToken verifies a given Google OAuth2 Token
func VerifyToken(token string) (User, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", googleVerifyURL, nil)
	req.Header.Add("Access_token", token)
	req.Header.Add("Authorization", "OAuth "+token)

	if err != nil {
		return User{}, err
	}

	resp, err := client.Do(req)

	if err != nil {
		return User{}, err
	}

	if resp.StatusCode != 200 {
		return User{}, errors.New("Invalid Token")
	}

	decoder := json.NewDecoder(resp.Body)
	var user User
	err = decoder.Decode(&user)

	if err != nil {
		return User{}, err
	}

	return user, err
}
