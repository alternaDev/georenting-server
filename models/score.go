package models

import "time"

// Score is a score for a geohash
type Score struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	GeoHash   string `gorm:"unique_index;primary_key"`
	LastVisit int64
	Score     float64
}
