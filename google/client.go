package google

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/alternaDev/georenting-server/models"
)

const (
	googleVerifyURL   = "https://www.googleapis.com/oauth2/v2/userinfo"
	googleGCMGroupURL = "https://android.googleapis.com/gcm/notification"
)

var (
	googleAPIKey    = os.Getenv("GOOGLE_API_KEY")
	googleProjectID = os.Getenv("GOOGLE_PROJECT_ID")
)

// User represents the data for a given Google User
type User struct {
	GoogleID      string `json:"id"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"FamilyName"`
	GooglePlusURL string `json:"link"`
	AvatarURL     string `json:"picture"`
	Gender        string `json:"gender"`
	Locale        string `json:"locale"`
}

type gcmGroupRequest struct {
	Operation           string   `json:"operation"`
	NotificationKey     string   `json:"notification_key"`
	NotificationKeyName string   `json:"notification_key_name"`
	RegistrationIDs     []string `json:"registration_ids"`
}

type gcmGroupResponse struct {
	NotificationKey string `json:"notification_key"`
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

func sendGCMGroupRequest(data gcmGroupRequest) (gcmGroupResponse, error) {
	httpClient := &http.Client{}

	bytes, err := json.Marshal(data)

	req, err := http.NewRequest("POST", googleGCMGroupURL, strings.NewReader(string(bytes)))
	req.Header.Add("Authorization", "key="+googleAPIKey)
	req.Header.Add("project_id", googleProjectID)

	resp, err := httpClient.Do(req)
	defer resp.Body.Close()

	if err != nil {
		return gcmGroupResponse{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)

	var response gcmGroupResponse
	json.Unmarshal(respBody, &response)

	return response, nil
}

// CreateDeviceGroup creates a new Device group on Google Cloud Messaging
func CreateDeviceGroup(firstID string, user models.User) error {
	response, err := sendGCMGroupRequest(gcmGroupRequest{
		Operation:           "create",
		NotificationKeyName: "GeoRenting-" + user.Name,
		RegistrationIDs:     []string{firstID},
	})

	if err != nil {
		return err
	}

	user.GCMNotificationID = response.NotificationKey

	return nil
}

// AddDeviceToGroup adds a device to a device group.
func AddDeviceToGroup(deviceID string, user models.User) error {
	_, err := sendGCMGroupRequest(gcmGroupRequest{
		Operation:           "add",
		NotificationKeyName: "GeoRenting-" + user.Name,
		NotificationKey:     user.GCMNotificationID,
		RegistrationIDs:     []string{deviceID},
	})

	if err != nil {
		return err
	}

	return nil
}

// RemoveDeviceFromGroup removes a device from a group.
func RemoveDeviceFromGroup(deviceID string, user models.User) error {
	_, err := sendGCMGroupRequest(gcmGroupRequest{
		Operation:           "remove",
		NotificationKeyName: "GeoRenting-" + user.Name,
		NotificationKey:     user.GCMNotificationID,
		RegistrationIDs:     []string{deviceID},
	})

	if err != nil {
		return err
	}

	return nil
}
