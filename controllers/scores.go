package controllers

import (
	"encoding/json"
	"log"
	"math"
	"net/http"

	"github.com/alternaDev/geomodel"
	"github.com/alternaDev/georenting-server/models"
)

type heatmapItemResponse struct {
	Latitude  float64
	Longitude float64
	Score     float64
}

// GetHeatmapHandler GET /scores/heatmap
// TODO: Cache this in redis.
func GetHeatmapHandler(w http.ResponseWriter, r *http.Request) {
	scores, err := models.FindAllScores()

	if err != nil {
		log.Printf("Error while fetching scores: %s", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]heatmapItemResponse, len(*scores))
	for i := range *scores {
		lat, lon := geomodel.DecodeGeoHash((*scores)[i].GeoHash)
		response[i].Latitude = lat
		response[i].Longitude = lon
		response[i].Score = (*scores)[i].Score
		if response[i].Score == math.NaN() {
			response[i].Score = -1
		}
	}

	bytes, err := json.Marshal(&response)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}
