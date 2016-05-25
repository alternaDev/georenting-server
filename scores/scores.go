package scores

import (
	"log"
	"math"

	geomodel "github.com/alternaDev/geomodel"
	models "github.com/alternaDev/georenting-server/models"
)

const (
	geoHashResolution                  = 6
	geoFenceBasePrice                  = 100
	magicalGeoRentingConstant          = 2
	secondaryMagicalGeoRentingConstant = 2.0
	geoFenceRentBasePrice              = 10.0
)

func RecordVisit(lat float64, lon float64, now int64) error {
	geoHash := geomodel.GeoCell(lat, lon, geoHashResolution)

	score := &models.Score{}
	err := models.DB.Where(models.Score{GeoHash: geoHash}).FirstOrInit(&score).Error

	if err != nil {
		return err
	}

	err = CalculateScore(score, now)

	if err != nil {
		return err
	}

	score.LastVisit = now

	err = models.DB.Save(&score).Error

	if err != nil {
		return err
	}

	return nil
}

func CalculateScore(score *models.Score, now int64) error {
	var tSum int64
	err := models.DB.Raw("SELECT SUM(? - last_visit) AS tsum FROM scores", now).Row().Scan(&tSum)

	if err != nil {
		return err
	}
	log.Printf("Sum: %d", tSum)

	var count int64
	err = models.DB.Model(&models.Score{}).Count(&count).Error

	if err != nil {
		return err
	}
	count = int64(math.Max(float64(count), 1))
	tAvg := float64((1.0 / float64(count)) * float64(tSum))
	fraction := tAvg / float64(now-score.LastVisit)
	logN := math.Log(fraction)
	score.Score = math.Max(0, math.Max(score.Score, 0)+logN)
	return models.DB.Save(&score).Error
}

func GetGeoFencePrice(lat float64, lon float64) (float64, error) {
	geoHash := geomodel.GeoCell(lat, lon, geoHashResolution)

	score := &models.Score{}
	err := models.DB.Where(models.Score{GeoHash: geoHash}).FirstOrInit(&score).Error

	if err != nil {
		return 0, err
	}

	return math.Pow(score.Score+1.0, 1.0/magicalGeoRentingConstant) * geoFenceBasePrice, nil
}

func GetGeoFenceRent(f *models.Fence) float64 {
	return float64(f.RentMultiplier)*geoFenceRentBasePrice + (((float64(f.RentMultiplier) - 1.0) * (float64(f.RentMultiplier) - 1.0)) / secondaryMagicalGeoRentingConstant)
}
