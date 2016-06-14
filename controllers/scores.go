package controllers

import (
  "net/http"
  "log"
  "encoding/json"
  "github.com/alternaDev/georenting-server/models"
  "github.com/alternaDev/geomodel"
)

type heatmapItemResponse struct {
  Latitude float64
  Longitude float64
  Score float64
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
  }


  bytes, err := json.Marshal(&response)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}
