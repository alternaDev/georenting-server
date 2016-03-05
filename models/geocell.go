package models

import "github.com/jinzhu/gorm"

// GeoCell is a geocell associated to a fence
type GeoCell struct {
	gorm.Model
	Value   string
	FenceID uint
}
