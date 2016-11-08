package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/alternaDev/georenting-server/models"
)

type upgradesResponse struct {
	Radius    []int     `json:"radius"`
	Rent      []float64 `json:"rent"`
	MaxTTL    int       `json:"max_ttl"`
	MinTTL    int       `json:"min_ttl"`
	MaxRadius int       `json:"max_radius"`
	MinRadius int       `json:"min_radius"`
}

// UpgradesHandler handles GET /application/upgrades
func UpgradesHandler(w http.ResponseWriter, r *http.Request) {

	data := upgradesResponse{
		Radius:    models.UpgradeTypesRadius[:],
		Rent:      models.UpgradeTypesRent[:],
		MaxTTL:    models.FenceMaxTTL,
		MinTTL:    models.FenceMinTTL,
		MaxRadius: models.FenceMaxRadius,
		MinRadius: models.FenceMinRadius,
	}

	bytes, err := json.Marshal(&data)

	if err != nil {
		InternalServerError(err, w)
		return
	}

	w.Write(bytes)
	return
}
