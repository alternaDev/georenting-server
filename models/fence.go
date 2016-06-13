package models

import (
	"time"

	"github.com/alternaDev/georenting-server/maths"
)

var (
	// UpgradeTypesRadius holds the possible Upgrade Types for Radius.
	UpgradeTypesRadius = [...]int{100, 150, 200, 250, 300, 350, 400}
	// UpgradeTypesRent holds the possible rent multipliers.
	UpgradeTypesRent = [...]float64{1, 1.5, 2, 2.5, 3, 3.5, 4}
	// FenceMaxTTL holds the maximum possible TTL of a fence.
	FenceMaxTTL = 60 * 60 * 24 * 7 // 7 days
	// FenceMinRadius holds the minimum radius of a fence.
	FenceMinRadius = maths.Min(UpgradeTypesRadius[:])
	// FenceMaxRadius holds the maximum radius of a fence.
	FenceMaxRadius = maths.Max(UpgradeTypesRadius[:])
)

// Fence is a fence
type Fence struct {
	ID             uint `gorm:"primary_key"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	User           User      `json:"-"`
	UserID         uint      `json:"owner_id"`
	Lat            float64   `json:"center_lat"`
	Lon            float64   `json:"center_lon"`
	Radius         int       `json:"radius"`
	RentMultiplier float64   `json:"rent_multiplier"`
	TTL            int       `json:"ttl"`
	DiesAt         time.Time `json:"diesAt"`
	Name           string    `json:"name"`
}

func (f *Fence) Save() error {
	return DB.Save(&f).Error
}

func (f *Fence) Delete() error {
	return DB.Delete(&f).Error
}

func FindFencesByIDs(ids []int64) ([]Fence, error) {
	result := make([]Fence, len(ids))
	err := DB.Where(ids).Find(&result).Error
	return result, err
}

func FindFenceByID(id interface{}) (Fence, error, bool) {
	var fence Fence
	req := DB.Preload("User").Find(&fence, id)

	return fence, req.Error, req.RecordNotFound()
}
