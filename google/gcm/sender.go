package gcm

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	gcmSendURL = "https://fcm.googleapis.com/fcm/"
)

type sendMessageResponse struct {
	Success               int64    `json:"success"`
	Failure               int64    `json:"failure"`
	FailedRegistrationIDs []string `json:"failed_registration_ids"`
}

// SendToGroup sends a GCM Message to a Device Group.
func SendToGroup(msg *Message) error {
	httpClient := &http.Client{}

	data, err := json.Marshal(msg)

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", gcmSendURL, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "key="+googleAPIKey)
	req.Header.Add("project_id", googleProjectID)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		var response sendMessageResponse
		err = json.Unmarshal(body, &response)

		if err != nil {
			return err
		}

		time.Sleep(500)

		for _, id := range response.FailedRegistrationIDs {
			msg.To = id
			req, err := http.NewRequest("POST", gcmSendURL, bytes.NewBuffer(data))
			if err != nil {
				return err
			}
			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Authorization", "key="+googleAPIKey)
			req.Header.Add("project_id", googleProjectID)
		}
	}

	return nil
}
