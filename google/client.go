package google

import (
	"encoding/json"
	"errors"
	"net/http"
)

const (
	googleVerifyURL = "https://www.googleapis.com/plus/v1/people/me"
)

// User represents the data for a given Google User
type User struct {
	GoogleID      string     `json:"id"`
	Name          string     `json:"displayName"`
	GooglePlusURL string     `json:"url"`
	Avatar        Avatar     `json:"image"`
	Cover         Cover      `json:"cover"`
	Gender        string     `json:"gender"`
	Locale        string     `json:"language"`
}

// Avatar represents a users avatar
type Avatar struct {
	URL string `json:"url"`
}

// Cover represents a persons cover
type Cover struct {
	CoverPhoto CoverPhoto `json:"coverPhoto"`
}

// CoverPhoto represents a users coverphoto
type CoverPhoto struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
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
