package controllers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/alternaDev/geomodel"
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
	w.Write([]byte("{}"))
}

// GetFencesHandler GET /fences
func GetFencesHandler(w http.ResponseWriter, r *http.Request) {

	lat, err1 := strconv.ParseFloat(r.URL.Query().Get("latitude"), 64)
	lon, err2 := strconv.ParseFloat(r.URL.Query().Get("longitude"), 64)
	radius, err3 := strconv.ParseFloat(r.URL.Query().Get("radius"), 64)
	userID, err4 := strconv.ParseUint(r.URL.Query().Get("user"), 10, 8)

	if err1 == nil && err2 == nil && err3 == nil {
		var result = geomodel.ProximityFetch(lat, lon, 20, radius, func(cells []string) []geomodel.LocationCapable {
			var result []geomodel.LocationCapable = make([]geomodel.LocationCapable, 0)

			var geoCells []models.GeoCell

			models.DB.Where("Value in (?)", cells).Find(&geoCells)

			for i := range geoCells {
				var fence models.Fence
				models.DB.Model(geoCells[i]).Related(&fence)
				exists := false
				for j := range result {
					if result[j].Key() == fence.Key() {
						exists = true
						break
					}
				}
				if !exists {
					result = append(result, fence)
				}
			}
			return result
		}, 20)

		fences := make([]fenceResponse, len(result))
		for i := range result {
			f := result[i].(models.Fence)
			fences[i].ID = f.ID
			fences[i].Lat = f.Lat
			fences[i].Lon = f.Lon
			fences[i].Name = f.Name
			fences[i].Radius = f.Radius
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
	geoCells := geomodel.GeoCells(f.Lat, f.Lon, 20)

	f.GeoCells = make([]models.GeoCell, len(geoCells))
	for i := range geoCells {
		f.GeoCells[i].Value = geoCells[i]
	}

	// TODO: Check overlap with other fences.

	models.DB.Save(&f)

	bytes, err := json.Marshal(&f)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}
