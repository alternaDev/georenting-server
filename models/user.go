package models

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	GoogleID          string
	Fences            []Fence
	PrivateKey        string
	GCMNotificationID string
	Name              string
	AvatarURL         string
}
