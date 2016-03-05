package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/alternaDev/geomodel"
	"github.com/alternaDev/georenting-server/auth"
	"github.com/alternaDev/georenting-server/models"
)

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

	models.DB.Save(&f)

	bytes, err := json.Marshal(&f)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}
