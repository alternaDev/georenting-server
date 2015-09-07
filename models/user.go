package models

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	GoogleID          string  `json:"-"`
	Fences            []Fence `json:"fences"`
	PrivateKey        string  `sql:"size:4096" json:"-"`
	GCMNotificationID string  `json:"-"`
	Name              string  `json:"name"`
	AvatarURL         string  `json:"avatar_url"`
	CoverURL          string  `json:"cover_url"`
}
