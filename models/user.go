package models

import "time"

const (
	// LastKnownGeoHashResolution is the resolution for the geohash of the last known position.
	LastKnownGeoHashResolution = 5
)

// User is a user.
type User struct {
	ID                      uint `gorm:"primary_key"`
	CreatedAt               time.Time
	UpdatedAt               time.Time
	GoogleID                string  `json:"-" gorm:"index"`
	Fences                  []Fence `json:"fences"`
	PrivateKey              string  `sql:"size:4096" json:"-"`
	GCMNotificationID       string  `json:"-"`
	Name                    string  `json:"name"`
	AvatarURL               string  `json:"avatar_url" gorm:"-"`
	Balance                 float64 `json:"balance"`
	LastKnownGeoHash        string  `json:"-"`
	EarningsRentAllTime     float64 `json:"-"`
	ExpensesRentAllTime     float64 `json:"-"`
	ExpensesGeoFenceAllTime float64 `json:"-"`
}
