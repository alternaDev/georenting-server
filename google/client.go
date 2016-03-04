package google

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const (
	googleVerifyURL  = "https://www.googleapis.com/oauth2/v4/token"
	googleProfileURL = "https://www.googleapis.com/plus/v1/people/me"
)

// TokenInfoResponse Gets the Token Info from da server.
type TokenInfoResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// User represents the data for a given Google User
type User struct {
	GoogleID      string `json:"id"`
	Name          string `json:"displayName"`
	GooglePlusURL string `json:"url"`
	Avatar        Avatar `json:"image"`
	Cover         Cover  `json:"cover"`
	Gender        string `json:"gender"`
	Locale        string `json:"language"`
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

	data := url.Values{}
	data.Set("client_id", os.Getenv("GOOGLE_CLIENT_ID"))
	data.Add("client_secret", os.Getenv("GOOGLE_CLIENT_SECRET"))
	data.Add("code", token)
	data.Add("grant_type", "authorization_code")

	req, err := http.NewRequest("POST", googleVerifyURL, bytes.NewBufferString(data.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	//req.Header.Add("Access_token", token)
	//req.Header.Add("Authorization", "OAuth "+token)

	if err != nil {
		return User{}, err
	}

	resp, err := client.Do(req)

	if err != nil {
		return User{}, err
	}

	if resp.StatusCode != 200 {
		return User{}, errors.New("Invalid Token.")
	}

	decoder := json.NewDecoder(resp.Body)
	var response TokenInfoResponse
	err = decoder.Decode(&response)

	if err != nil {
		return User{}, err
	}

	req, err = http.NewRequest("GET", googleProfileURL, nil)
	req.Header.Add("Access_token", response.AccessToken)
	req.Header.Add("Authorization", "OAuth "+response.AccessToken)

	if err != nil {
		return User{}, err
	}

	resp, err = client.Do(req)

	if err != nil {
		return User{}, err
	}

	if resp.StatusCode != 200 {
		return User{}, errors.New("Invalid User")
	}

	var user User
	err = decoder.Decode(&user)

	if err != nil {
		return User{}, err
	}

	return user, err
}
