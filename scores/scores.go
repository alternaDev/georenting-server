package scores

import (
	"errors"
	"math"

	geomodel "github.com/alternaDev/geomodel"
	models "github.com/alternaDev/georenting-server/models"
)

const (
	geoHashResolution                  = 6
	geoFenceBasePrice                  = 100.0
	magicalGeoRentingConstant          = 2
	secondaryMagicalGeoRentingConstant = 2.0
	geoFenceRentBasePrice              = 10.0
)

var (
	InitialBalance = 0.0
)

func init() {
	balance, err := GetGeoFencePriceForScore(0, models.FenceMinTTL, 1, 0)
	if err != nil {
		panic(err)
	}
	InitialBalance = balance * 5 // move this to a constant later.
}

// RecordVisit calculates the new score of a geofence after a visit.
func RecordVisit(lat float64, lon float64, now int64) error {
	geoHash := geomodel.GeoCell(lat, lon, geoHashResolution)

	score, err := models.FindScoreByGeoHashOrInit(geoHash)

	if err != nil {
		return errors.New("Failed to find score: " + err.Error())
	}

	err = CalculateScore(score, now)

	if err != nil {
		return errors.New("Failed to Save while calclating score: " + err.Error())
	}

	score.LastVisit = now

	err = score.Save()

	if err != nil {
		return errors.New("Failed to save: " + err.Error())
	}

	return nil
}

// CalculateScore calculates a geofence score.
func CalculateScore(score models.Score, now int64) error {
	var tSum int64
	err := models.DB.Get(&tSum, "SELECT SUM($1 - last_visit) AS tsum FROM scores", now)

	if err != nil {
		return err
	}

	count, err := models.CountScores()

	if err != nil {
		return errors.New("Failed to count scores: " + err.Error())
	}

	count = int64(math.Max(float64(count), 1))
	tAvg := float64((1.0 / float64(count)) * float64(tSum))
	fraction := tAvg / float64(now-score.LastVisit)
	logN := math.Log(fraction)
	score.Score = math.Max(0, math.Max(score.Score, 0)+logN)

	if score.Score == math.NaN() {
		return errors.New("NaN Error!")
	}

	return score.Save()
}

// GetGeoFencePrice returns the price of a geofence depending on the upgrade status and current score.
func GetGeoFencePrice(lat float64, lon float64, ttl int, rentMultiplier float64, radiusIndex int) (float64, error) {
	geoHash := geomodel.GeoCell(lat, lon, geoHashResolution)

	score, err := models.FindScoreByGeoHashOrInit(geoHash)

	if err != nil {
		return 0, err
	}

	return GetGeoFencePriceForScore(score.Score, ttl, rentMultiplier, radiusIndex)
}

// GetGeoFencePriceForScore returns the price of a geofence depending on the upgrade status and current score.
func GetGeoFencePriceForScore(score float64, ttl int, rentMultiplier float64, radiusIndex int) (float64, error) {
	price := math.Pow(score+1.0, 1.0/magicalGeoRentingConstant) * geoFenceBasePrice

	ttlMultiplier := ((float64(ttl) / float64(models.FenceMaxTTL)) + 1.0)
	radiusMultiplier := (((float64(radiusIndex + 1)) / (float64(len(models.UpgradeTypesRadius)))) + 1.0)

	value := rentMultiplier * ttlMultiplier * radiusMultiplier * price
	return value, nil
}

// GetGeoFenceRent returns the rent of a geofence.
func GetGeoFenceRent(f models.Fence) float64 {
	return float64(f.RentMultiplier)*geoFenceRentBasePrice + (((float64(f.RentMultiplier) - 1.0) * (float64(f.RentMultiplier) - 1.0)) / secondaryMagicalGeoRentingConstant)
}
