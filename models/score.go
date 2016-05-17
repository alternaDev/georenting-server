package models

import (
	"github.com/jinzhu/gorm"
)

// Score is a score for a geohash
type Score struct {
	gorm.Model
	GeoHash     string `gorm:"unique_index"`
  LastVisit   int64
  Score       float64
}
