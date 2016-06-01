package activity

import (
	"encoding/json"
	"time"

	redis "github.com/alternaDev/georenting-server/models/redis"
)

type ownFenceVisitedActivity struct {
	Verb        string  `json:"verb"`
	VisitorName string  `json:"visitorName"`
	VisitorID   uint    `json:"visitorId"`
	FenceName   string  `json:"fenceName"`
	FenceID     uint    `json:"fenceId"`
	Rent        float64 `json:"rent"`
	Time        int32   `json:"time"`
}

type foreignFenceVisitedActivity struct {
	Verb      string  `json:"verb"`
	OwnerName string  `json:"ownerName"`
	OwnerID   uint    `json:"ownerId"`
	FenceName string  `json:"fenceName"`
	FenceID   uint    `json:"fenceId"`
	Rent      float64 `json:"rent"`
	Time      int32   `json:"time"`
}

type fenceExpiredActivity struct {
	Verb      string `json:"verb"`
	FenceID   uint   `json:"fenceId"`
	FenceName string `json:"fenceName"`
	Time      int32  `json:"time"`
}

// AddOwnFenceVisitedActivity adds the activity to the stream of the owner.
func AddOwnFenceVisitedActivity(ownerID uint, visitorName string, visitorID uint, fenceName string, fenceID uint, rent float64) error {
	now := int32(time.Now().Unix())

	bytes, err := json.Marshal(ownFenceVisitedActivity{Verb: "ownFenceVisited",
		VisitorName: visitorName,
		VisitorID:   visitorID,
		FenceName:   fenceName,
		FenceID:     fenceID,
		Rent:        rent,
		Time:        now})

	if err != nil {
		return err
	}

	return redis.AddActivity(ownerID, float64(now), string(bytes[:]))
}

// AddForeignVisitedActivity adds the activity to the stream of the owner.
func AddForeignVisitedActivity(visitorID uint, ownerName string, ownerID uint, fenceName string, fenceID uint, rent float64) error {
	now := int32(time.Now().Unix())

	bytes, err := json.Marshal(foreignFenceVisitedActivity{Verb: "foreignFenceVisited",
		OwnerName: ownerName,
		OwnerID:   ownerID,
		FenceName: fenceName,
		FenceID:   fenceID,
		Rent:      rent,
		Time:      now})

	if err != nil {
		return err
	}

	return redis.AddActivity(visitorID, float64(now), string(bytes[:]))
}

// AddFenceExpiredActivity adds the activity to the stream of the owner.
func AddFenceExpiredActivity(ownerID uint, fenceID uint, fenceName string) error {
	now := int32(time.Now().Unix())

	bytes, err := json.Marshal(fenceExpiredActivity{Verb: "fenceExpired",
		FenceID:   fenceID,
		FenceName: fenceName,
		Time:      now})

	if err != nil {
		return err
	}

	return redis.AddActivity(ownerID, float64(now), string(bytes[:]))
}

// GetActivities returns all activities from the specified user in a timerange.
func GetActivities(userID uint, start int64, end int64) ([]string, error) {
	return redis.GetActivities(userID, start, end)
}
