package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alternaDev/georenting-server/auth"
	"github.com/alternaDev/georenting-server/google/gcm"
	"github.com/alternaDev/georenting-server/jobs"
)

type gcmID struct {
	GCMID string `json:"gcm_id"`
}

type gcmNotificationKeyResponse struct {
	NotificationKey string `json:"gcm_notification_key"`
}

// GCMAddHandler POST /users/me/gcm
func GCMAddHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.ValidateSession(r)

	if err != nil {
		http.Error(w, "Invalid Session token. "+err.Error(), http.StatusForbidden)
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
			err = gcm.CreateDeviceGroup(b.GCMID, user)

			if err != nil {
				InternalServerError(err, w)
				return
			}
		}
	} else {
		err = gcm.CreateDeviceGroup(b.GCMID, user)

		if err != nil {
			InternalServerError(err, w)
			return
		}
	}

	err = jobs.QueueSendGcmRequest(gcm.NewMessage(map[string]interface{}{"type": "sync"}, user.GCMNotificationID))

	if err != nil {
		InternalServerError(err, w)
		return
	}

	bytes, err := json.Marshal(gcmNotificationKeyResponse{NotificationKey: user.GCMNotificationID})

	if err != nil {
		InternalServerError(err, w)
		return
	}

	w.Write(bytes)
}

// GCMRemoveHandler DELETE /users/me/gcm
func GCMRemoveHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.ValidateSession(r)

	if err != nil {
		http.Error(w, "Invalid Session token. "+err.Error(), http.StatusForbidden)
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
		err = gcm.RemoveDeviceFromGroup(b.GCMID, user)

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
	} else {
		http.Error(w, "Could not remove Token from nonexisting Group.", http.StatusUnauthorized)
		return
	}

	fmt.Fprintf(w, "{}")
}
