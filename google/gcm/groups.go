package gcm

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/alternaDev/georenting-server/models"
)

const (
	googleGCMGroupURL = "https://android.googleapis.com/gcm/notification"
)

type gcmGroupRequest struct {
	Operation           string   `json:"operation"`
	NotificationKey     string   `json:"notification_key"`
	NotificationKeyName string   `json:"notification_key_name"`
	RegistrationIDs     []string `json:"registration_ids"`
}

type gcmGroupResponse struct {
	NotificationKey string `json:"notification_key"`
	Error           string `json:"error"`
}

func sendGCMGroupRequest(data gcmGroupRequest) (gcmGroupResponse, error) {
	httpClient := &http.Client{}

	bytes, err := json.Marshal(data)

	req, err := http.NewRequest("POST", googleGCMGroupURL, strings.NewReader(string(bytes)))
	req.Header.Add("Content-Type", "application/json")
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

	if response.Error != "" {
		return errors.New(response.Error)
	}

	user.GCMNotificationID = response.NotificationKey
	models.DB.Save(&user)

	return nil
}

// AddDeviceToGroup adds a device to a device group.
func AddDeviceToGroup(deviceID string, user models.User) error {
	response, err := sendGCMGroupRequest(gcmGroupRequest{
		Operation:           "add",
		NotificationKeyName: "GeoRenting-" + user.Name,
		NotificationKey:     user.GCMNotificationID,
		RegistrationIDs:     []string{deviceID},
	})

	if err != nil {
		return err
	}

	if response.Error != "" {
		return errors.New(response.Error)
	}

	return nil
}

// RemoveDeviceFromGroup removes a device from a group.
func RemoveDeviceFromGroup(deviceID string, user models.User) error {
	response, err := sendGCMGroupRequest(gcmGroupRequest{
		Operation:           "remove",
		NotificationKeyName: "GeoRenting-" + user.Name,
		NotificationKey:     user.GCMNotificationID,
		RegistrationIDs:     []string{deviceID},
	})

	if err != nil {
		return err
	}

	if response.Error != "" {
		return errors.New(response.Error)
	}

	return nil
}
