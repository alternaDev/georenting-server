package activity

import (
	"encoding/json"
	"fmt"
	"time"

	models "github.com/alternaDev/georenting-server/models"
	redis "gopkg.in/redis.v3"
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

	return models.RedisInstance.ZAdd(fmt.Sprintf("%v", ownerID), redis.Z{Score: float64(now), Member: string(bytes[:])}).Err()
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

	return models.RedisInstance.ZAdd(fmt.Sprintf("%v", visitorID), redis.Z{Score: float64(now), Member: string(bytes[:])}).Err()
}

/*func GetActivities(userID uint, start int32, end int32) {
	models.RedisInstance.ZRevRangeByScore(userID, start, end)
}*/
