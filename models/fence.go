package models

import "github.com/jinzhu/gorm"

// Fence is a fence
type Fence struct {
	gorm.Model
	Owner    User
	OwnerID  int     `sql:"index"`
	Lat      float64 `json:"centerLat"`
	Lon      float64 `json:"centerLon"`
	Radius   int
	Name     string   `json:"name"`
	GeoCells []string `sql:"type:jsonb"`
}
