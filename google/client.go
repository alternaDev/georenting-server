package google

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
)

const (
	googleVerifyURL  = "https://www.googleapis.com/oauth2/v3/tokeninfo"
	googleProfileURL = "https://www.googleapis.com/plus/v1/people/me"
)

// TokenInfoResponse Gets the Token Info from da server.
type TokenInfoResponse struct {
	Issuer     string `json:"iss"`
	Audit      string `json:"aud"`
	Sub        string `json:"sub"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Avatar     string `json:"picture"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Locale     string `json:"locale"`
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

	req, err := http.NewRequest("GET", googleVerifyURL+"?id_token="+token, nil)

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
		return User{}, errors.New("Invalid Token")
	}

	decoder := json.NewDecoder(resp.Body)
	var response TokenInfoResponse
	err = decoder.Decode(&response)

	if err != nil {
		return User{}, err
	}

	if response.Audit != os.Getenv("GOOGLE_CLIENT_ID") {
		return User{}, errors.New("Invalid Client.")
	}

	// TODO: Check whether the Token was valid.

	req, err = http.NewRequest("GET", googleProfileURL, nil)
	req.Header.Add("Access_token", token)
	req.Header.Add("Authorization", "OAuth "+token)

	if err != nil {
		return User{}, err
	}

	resp, err = client.Do(req)

	if err != nil {
		return User{}, err
	}

	if resp.StatusCode != 200 {
		return User{}, errors.New("Invalid Token")
	}

	var user User
	err = decoder.Decode(&user)

	if err != nil {
		return User{}, err
	}

	return user, err
}
