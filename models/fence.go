package models

import "github.com/jinzhu/gorm"

type Fence struct {
	gorm.Model
	lat    float64
	lon    float64
	radius int
	name   string
}
