package controllers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/alternaDev/geomodel"
	"github.com/alternaDev/georenting-server/models"
	ourRedis "github.com/alternaDev/georenting-server/models/redis"
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

	response, err := ourRedis.RedisInstance.Get(redisKeyAllScoreCache).Result()
	if err != nil || response == "\"\"" {
		scores, err := models.FindAllScores()

		if err != nil {
			InternalServerError(err, w)
			return
		}

		r := make([]heatmapItemResponse, len(scores))
		for i := range scores {
			lat, lon := geomodel.DecodeGeoHash(scores[i].GeoHash)
			r[i].Latitude = lat
			if r[i].Latitude == math.NaN() {
				r[i].Latitude = 0
			}
			r[i].Longitude = lon
			if r[i].Longitude == math.NaN() {
				r[i].Longitude = 0
			}
			r[i].Score = scores[i].Score
			if r[i].Score == math.NaN() {
				r[i].Score = -1
			}
		}

		bytes, err := json.Marshal(&r)

		if err != nil {
			InternalServerError(err, w)
			return
		}

		response = string(bytes)

		ourRedis.RedisInstance.Set(redisKeyAllScoreCache, response, redisAllScoreCacheTTL*time.Second).Err()
		if err != nil {
			log.Errorf("Error while caching scores: %s", err)
		}
	}

	fmt.Fprintf(w, response)
}
