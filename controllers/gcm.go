package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/alternaDev/georenting-server/auth"
	"github.com/alternaDev/georenting-server/google/gcm"
)

type gcmID struct {
	GCMID string `json:"gcm_id"`
}

type gcmNotificationKeyResponse struct {
	NotificationKey string `json:"gcm_notification_key"`
}

// GCMAddFunc POST /users/me/gcm
func GCMAddFunc(w http.ResponseWriter, r *http.Request) {

	user, err := auth.ValidateSession(r)

	if err != nil {
		http.Error(w, "Invalid Session token. "+err.Error(), http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var b gcmID
	err = decoder.Decode(&b)

	if err != nil {
		http.Error(w, "Invalid Body.", http.StatusBadRequest)
		return
	}

	if user.GCMNotificationID != "" {
		err = gcm.AddDeviceToGroup(b.GCMID, user)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		err = gcm.CreateDeviceGroup(b.GCMID, user)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	err = gcm.SendToGroup(gcm.NewMessage(map[string]interface{}{"type": "sync"}, user.GCMNotificationID))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(gcmNotificationKeyResponse{NotificationKey: user.GCMNotificationID})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}
