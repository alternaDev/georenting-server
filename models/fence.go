package models

import (
	"strconv"

	"github.com/jinzhu/gorm"
)

// Fence is a fence
type Fence struct {
	gorm.Model
	User     User
	UserID   uint
	Lat      float64 `json:"centerLat"`
	Lon      float64 `json:"centerLon"`
	Radius   int
	Name     string `json:"name"`
	GeoCells []GeoCell
}

func (p Fence) Latitude() float64 {
	return p.Lat
}

func (p Fence) Longitude() float64 {
	return p.Lon
}

func (p Fence) Key() string {
	return strconv.Itoa(int(p.ID))
}

func (p Fence) Geocells() []string {
	var cells = make([]string, len(p.GeoCells))

	for i := range p.GeoCells {
		cells[i] = p.GeoCells[i].Value
	}

	return cells
}
