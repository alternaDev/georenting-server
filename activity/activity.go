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

	return addActivity(ownerID, float64(now), string(bytes[:]))
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

	return addActivity(visitorID, float64(now), string(bytes[:]))
}

func addActivity(userID uint, score float64, data string) error {
	return models.RedisInstance.ZAdd(fmt.Sprintf("%v", userID), redis.Z{Score: score, Member: data}).Err()
}

func GetActivities(userID uint, start int64, end int64) ([]string, error) {
	return models.RedisInstance.ZRevRangeByScore(fmt.Sprintf("%v", userID), redis.ZRangeByScore{Min: fmt.Sprintf("%v", start), Max: fmt.Sprintf("%v", end)}).Result()
}
