package scores

import (
  "time"
  "math"

  geomodel "github.com/alternaDev/geomodel"
  models "github.com/alternaDev/georenting-server/models"
)

const (
  geoHashResolution = 6
)

func RecordVisit(lat float64, lon float64) (error) {
  geoHash := geomodel.GeoCell(lat, lon, geoHashResolution)

  score := &models.Score{}
  err := models.DB.Where(models.Score{GeoHash: geoHash}).FirstOrInit(&score).Error

  if err != nil {
    return err
  }

  score.LastVisit = time.Now().Unix()

  err = models.DB.Save(&score).Error

  if err != nil {
    return err
  }

  return CalculateScore(score)
}

func CalculateScore(score *models.Score) (error) {
  type SumResult struct {
    tSum int64
  }
  var sumResult SumResult
  err := models.DB.Raw("SELECT SUM(? - LastVisit) AS tSum FROM Score", time.Now().Unix()).Scan(&sumResult).Error

  if err != nil {
    return err
  }
  tSum := sumResult.tSum

  var count int64
  err = models.DB.Model(&models.Score{}).Count(&count).Error

  if err != nil {
    return err
  }

  tAvg := float64((1 / count) * tSum)

  score.Score = math.Max(0, score.Score + math.Log(tAvg / float64(time.Now().Unix() - score.LastVisit)))
  return models.DB.Save(&score).Error
}
