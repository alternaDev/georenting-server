package activity

import (
	"encoding/json"
	"time"

	redis "github.com/alternaDev/georenting-server/models/redis"
)

type ownFenceVisitedActivity struct {
	Verb        string  `json:"verb"`
	VisitorName string  `json:"visitorName"`
	VisitorID   int     `json:"visitorId"`
	FenceName   string  `json:"fenceName"`
	FenceID     int     `json:"fenceId"`
	Rent        float64 `json:"rent"`
	Time        int32   `json:"time"`
}

type foreignFenceVisitedActivity struct {
	Verb      string  `json:"verb"`
	OwnerName string  `json:"ownerName"`
	OwnerID   int     `json:"ownerId"`
	FenceName string  `json:"fenceName"`
	FenceID   int     `json:"fenceId"`
	Rent      float64 `json:"rent"`
	Time      int32   `json:"time"`
}

type fenceExpiredActivity struct {
	Verb      string `json:"verb"`
	FenceID   int    `json:"fenceId"`
	FenceName string `json:"fenceName"`
	Time      int32  `json:"time"`
}

// AddOwnFenceVisitedActivity adds the activity to the stream of the owner.
func AddOwnFenceVisitedActivity(ownerID int, visitorName string, visitorID int, fenceName string, fenceID int, rent float64) error {
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
func AddForeignVisitedActivity(visitorID int, ownerName string, ownerID int, fenceName string, fenceID int, rent float64) error {
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
func AddFenceExpiredActivity(ownerID int, fenceID int, fenceName string) error {
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
func GetActivities(userID int, start int64, end int64) ([]string, error) {
	return redis.GetActivities(userID, start, end)
}
