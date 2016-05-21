package scores

import (
  "time"
  "math"
  "log"

  geomodel "github.com/alternaDev/geomodel"
  models "github.com/alternaDev/georenting-server/models"
)

const (
  geoHashResolution = 6
  geoFenceBasePrice = 100
)

func RecordVisit(lat float64, lon float64) (error) {
  geoHash := geomodel.GeoCell(lat, lon, geoHashResolution)

  score := &models.Score{}
  err := models.DB.Where(models.Score{GeoHash: geoHash}).FirstOrInit(&score).Error

  if err != nil {
    return err
  }

  err = CalculateScore(score)

  if err != nil {
    return err
  }

  score.LastVisit = time.Now().Unix()

  err = models.DB.Save(&score).Error

  if err != nil {
    return err
  }

  return nil
}

func CalculateScore(score *models.Score) (error) {
  var tSum int64
  err := models.DB.Raw("SELECT SUM(? - last_visit) AS tsum FROM scores", time.Now().Unix()).Row().Scan(&tSum)

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
  fraction := tAvg / float64(time.Now().Unix() - score.LastVisit)
  logN := math.Log(fraction)
  score.Score = math.Max(0, math.Max(score.Score, 0) + logN)
  return models.DB.Save(&score).Error
}

func GetGeoFencePrice(lat float64, lon float64) (float64, error) {
  geoHash := geomodel.GeoCell(lat, lon, geoHashResolution)

  score := &models.Score{}
  err := models.DB.Where(models.Score{GeoHash: geoHash}).FirstOrInit(&score).Error

  if err != nil {
    return 0, err
  }

  return math.Sqrt(score.Score + 1.0) * geoFenceBasePrice, nil
}
