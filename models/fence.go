package models

import (
	"github.com/alternaDev/georenting-server/maths"
	"github.com/jinzhu/gorm"
)

var (
	UpgradeTypesRadius = [...]int{100, 150, 200, 250, 300, 350, 400}
	UpgradeTypesRent   = [...]float32{1, 1.5, 2, 2.5, 3, 3.5, 4}
	UpgradeTypesTTL    = [...]int{3, 5, 7}
	FenceMinRadius     = maths.Min(UpgradeTypesRadius[:])
	FenceMaxRadius     = maths.Max(UpgradeTypesRadius[:])
)

// Fence is a fence
type Fence struct {
	gorm.Model
	User   User    `json:"-"`
	UserID uint    `json:"ownerId"`
	Lat    float64 `json:"centerLat"`
	Lon    float64 `json:"centerLon"`
	Radius int     `json:"radius"`
	Name   string  `json:"name"`
}
