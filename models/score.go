package models

import "time"
import "log"

// Score is a score for a geohash
type Score struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	GeoHash   string `gorm:"unique_index;primary_key"`
	LastVisit int64
	Score     float64
}

func (s Score) Save() error {
	return DB.Save(&s).Error
}

func FindScoreByGeoHashOrInit(geoHash string) (*Score, error) {
	var result Score

	err := DB.Where(&Score{GeoHash: geoHash}).FirstOrCreate(&result).Error
	return &result, err
}

func FindAllScores() (*[]Score, error) {
	var result []Score
	err := DB.Find(&result).Error
	log.Printf("Found %d scores.", len(result))
	return &result, err
}

func CountScores() (int64, error) {
	var count int64
	err := DB.Model(&Score{}).Count(&count).Error
	return count, err
}
