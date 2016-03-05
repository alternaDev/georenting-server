package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/alternaDev/geomodel"
	"github.com/alternaDev/georenting-server/auth"
	"github.com/alternaDev/georenting-server/models"
)

// GetFencesHandler GET /fences
func GetFencesHandler(w http.ResponseWriter, r *http.Request) {
	lat, err1 := strconv.ParseFloat(r.URL.Query()["latitude"][0], 64)
	lon, err2 := strconv.ParseFloat(r.URL.Query()["longitude"][0], 64)
	radius, err3 := strconv.ParseFloat(r.URL.Query()["radius"][0], 64)

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

		fences := make([]models.Fence, len(result))
		for i := range result {
			fences[i] = result[i].(models.Fence)
		}

		bytes, err := json.Marshal(&fences)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(bytes)
	}
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

	f.Owner = user
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
