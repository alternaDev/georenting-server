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

func (u *User) Save() error {
	return DB.Save(&u).Error
}

func (u *User) GetFences() *[]Fence {
	var fences *[]Fence

	DB.Model(&u).Related(&fences)

	return fences
}

func FindUserByID(id interface{}) (*User, error) {
	var result User
	err := DB.First(&result, id).Error
	return &result, err
}

func FindUsersByLastKnownGeoHash(hash string) ([]User, error) {
	var users []User
	err := DB.Where(User{LastKnownGeoHash: hash}).Find(&users).Error
	return users, err
}

func FindUserByGoogleIDOrInit(id string) (*User, error) {
	var result *User
	err := DB.FirstOrInit(&result, User{GoogleID: id}).Error
	return result, err
}

func CountUsersByName(name string) (int, error) {
	count := 0
	err := DB.Where(User{Name: name}).Count(&count).Error
	return count, err
}
