package models

import (
	"github.com/jinzhu/gorm"
)

// Fence is a fence
type Fence struct {
	gorm.Model
	User     User      `json:"-"`
	UserID   uint      `json:"ownerId"`
	Lat      float64   `json:"centerLat"`
	Lon      float64   `json:"centerLon"`
	Radius   int       `json:"radius"`
	Name     string    `json:"name"`
}
