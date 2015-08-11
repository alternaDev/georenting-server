package models

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	GoogleID          string
	Fences            []Fence
	PrivateKey        string `sql:"size:4096"`
	GCMNotificationID string
	Name              string
	AvatarURL         string
}
