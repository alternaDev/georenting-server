package controllers

import (
	"encoding/json"
	"log"
	"math"
	"net/http"

	"github.com/alternaDev/geomodel"
	"github.com/alternaDev/georenting-server/models"
	ourRedis "github.com/alternaDev/georenting-server/models/redis"
	"gopkg.in/redis.v3"

)

const (
	redisKeyAllScoreCache = "heatmap_scores"
	redisAllScoreCacheTTL = 60 * 60 * 60
)

type heatmapItemResponse struct {
	Latitude  float64
	Longitude float64
	Score     float64
}

// GetHeatmapHandler GET /scores/heatmap
func GetHeatmapHandler(w http.ResponseWriter, r *http.Request) {
	var response string

	response, err := ourRedis.RedisInstance.Get(redisKeyAllScoreCache).Result()

	if err == redis.Nil || response == "" {
		scores, err := models.FindAllScores()

		if err != nil {
			log.Printf("Error while fetching scores: %s", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		r := make([]heatmapItemResponse, len(*scores))
		for i := range *scores {
			lat, lon := geomodel.DecodeGeoHash((*scores)[i].GeoHash)
			r[i].Latitude = lat
			if r[i].Latitude == math.NaN() {
				r[i].Latitude = 0
			}
			r[i].Longitude = lon
			if r[i].Longitude == math.NaN() {
				r[i].Longitude = 0
			}
			r[i].Score = (*scores)[i].Score
			if r[i].Score == math.NaN() {
				r[i].Score = -1
			}
		}

		bytes, err := json.Marshal(&response)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response = string(bytes)

		ourRedis.RedisInstance.Set(redisKeyAllScoreCache, response, redisAllScoreCacheTTL)
	}


	w.Write([]byte(response))
}
