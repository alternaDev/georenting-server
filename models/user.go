package models

import "github.com/jinzhu/gorm"

const (
	// LastKnownGeoHashResolution is the resolution for the geohash of the last known position.
	LastKnownGeoHashResolution = 5
)

// User is a user.
type User struct {
	gorm.Model
	GoogleID                string  `json:"-"`
	Fences                  []Fence `json:"fences"`
	PrivateKey              string  `sql:"size:4096" json:"-"`
	GCMNotificationID       string  `json:"-"`
	Name                    string  `json:"name"`
	AvatarURL               string  `json:"avatar_url"`
	CoverURL                string  `json:"cover_url"`
	Balance                 float64 `json:"balance"`
	LastKnownGeoHash        string  `json:"-"`
	EarningsRentAllTime     float64 `json:"-"`
	ExpensesRentAllTime     float64 `json:"-"`
	ExpensesGeoFenceAllTime float64 `json:"-"`
}
