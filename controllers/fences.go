package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"log"
	"fmt"

	"github.com/alternaDev/georenting-server/activity"
	"github.com/alternaDev/georenting-server/auth"
	"github.com/alternaDev/georenting-server/google/gcm"
	"github.com/alternaDev/georenting-server/models"
	"github.com/gorilla/mux"
)

type fenceResponse struct {
	ID     uint    `json:"id"`
	Lat    float64 `json:"centerLat"`
	Lon    float64 `json:"centerLon"`
	Radius int     `json:"radius"`
	Name   string  `json:"name"`
	Owner  uint    `json:"owner"`
}

// VisitFenceHandler handles POST /fences/{fenceId}/visit
func VisitFenceHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.ValidateSession(r)

	if err != nil {
		http.Error(w, "Invalid Session token. "+err.Error(), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)

	fenceID, err := strconv.ParseUint(vars["fenceId"], 10, 8)
	if err != nil {
		http.Error(w, "Invalid Fence ID. "+err.Error(), http.StatusUnauthorized)
		return
	}

	var fence models.Fence

	models.DB.Preload("User").Find(&fence, fenceID)

	//TODO: Do money calculations and all those things.

	rent := 100.0

	// GCM
	err = gcm.SendToGroup(gcm.NewMessage(map[string]interface{}{"type": "onForeignFenceEntered", "fenceId": fence.ID, "fenceName": fence.Name, "ownerName": fence.User.Name}, user.GCMNotificationID))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = gcm.SendToGroup(gcm.NewMessage(map[string]interface{}{"type": "onOwnFenceEntered", "fenceId": fence.ID, "fenceName": fence.Name, "visitorName": user.Name}, fence.User.GCMNotificationID))

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Activity Stream
	err = activity.AddForeignVisitedActivity(user.ID, fence.User.Name, fence.User.ID, fence.Name, fence.ID, rent)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = activity.AddOwnFenceVisitedActivity(fence.User.ID, user.Name, user.ID, fence.Name, fence.ID, rent)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("{}"))
}

// GetFencesHandler GET /fences
func GetFencesHandler(w http.ResponseWriter, r *http.Request) {

	lat, err1 := strconv.ParseFloat(r.URL.Query().Get("latitude"), 64)
	lon, err2 := strconv.ParseFloat(r.URL.Query().Get("longitude"), 64)
	radius, err3 := strconv.ParseInt(r.URL.Query().Get("radius"), 10, 64)
	userID, err4 := strconv.ParseUint(r.URL.Query().Get("user"), 10, 8)

	if err1 == nil && err2 == nil && err3 == nil {

		ids, err := models.FindGeoFences(lat, lon, radius)
		if err != nil {
			log.Fatal(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		result := make([]models.Fence, len(ids))
		models.DB.Where(ids).Find(&result)


		fences := make([]fenceResponse, len(result))
		for i := range result {
			f := result[i]
			fences[i].ID = f.ID
			fences[i].Lat = f.Lat
			fences[i].Lon = f.Lon
			fences[i].Name = f.Name
			fences[i].Radius = f.Radius
			fences[i].Owner = f.UserID
		}

		bytes, err := json.Marshal(&fences)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(bytes)
		return
	}

	if err4 == nil {
		var user models.User
		models.DB.Preload("Fences").First(&user, userID)
		result := user.Fences

		fences := make([]fenceResponse, len(result))
		for i := range result {
			f := result[i]
			fences[i].ID = f.ID
			fences[i].Lat = f.Lat
			fences[i].Lon = f.Lon
			fences[i].Name = f.Name
			fences[i].Radius = f.Radius
			fences[i].Owner = f.UserID
		}

		bytes, err := json.Marshal(&fences)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(bytes)
		return
	}

	err := err1
	if err == nil {
		err = err2
	}
	if err == nil {
		err = err3
	}
	if err == nil {
		err = err4
	}
	if err == nil {
		err = errors.New("Please specify valid query options.")
	}

	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// CreateFenceHandler POST /fences
func CreateFenceHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.ValidateSession(r)

	if err != nil {
		http.Error(w, "Invalid Session token. "+err.Error(), http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var f models.Fence
	err = decoder.Decode(&f)

	f.User = user
	f.Radius = 100
	// TODO: Check overlap with other fences.

	models.DB.Save(&f)

	err = models.IndexGeoFence(&f)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	bytes, err := json.Marshal(&f)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}

// GetFenceHandler GET /fences/{fenceId}
func GetFenceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	fenceID, err := strconv.ParseUint(vars["fenceId"], 10, 8)
	if err != nil {
		http.Error(w, "Invalid Fence ID. "+err.Error(), http.StatusUnauthorized)
		return
	}

	var fence models.Fence

	notFound := models.DB.Find(&fence, fenceID).RecordNotFound()

	if notFound {
		http.Error(w, "GeoFence Not Found.", http.StatusNotFound)
		return
	}

	var f fenceResponse
	f.ID = fence.ID
	f.Lat = fence.Lat
	f.Lon = fence.Lon
	f.Name = fence.Name
	f.Radius = fence.Radius
	f.Owner = fence.UserID

	bytes, err := json.Marshal(&f)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}

// RemoveFenceHandler DELETE /fences/{fenceId}
func RemoveFenceHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.ValidateSession(r)

	if err != nil {
		http.Error(w, "Invalid Session token. "+err.Error(), http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)

	fenceID, err := strconv.ParseUint(vars["fenceId"], 10, 8)
	if err != nil {
		http.Error(w, "Invalid Fence ID. "+err.Error(), http.StatusUnauthorized)
		return
	}

	var fence models.Fence

	notFound := models.DB.Find(&fence, fenceID).RecordNotFound()

	if notFound {
		http.Error(w, "GeoFence Not Found.", http.StatusNotFound)
		return
	}

	if fence.UserID != user.ID {
		http.Error(w, "Unauthorized User.", http.StatusUnauthorized)
		return
	}


	err = models.DeleteGeoFence(&fence)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = models.DB.Delete(fence).Error

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "{}")
}
