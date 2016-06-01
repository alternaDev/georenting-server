package redis

import (
	"net/url"
	"os"
	"fmt"
	"time"
	"log"


	"gopkg.in/redis.v3"
)

// RedisInstance is a usable redis instance.
var RedisInstance *redis.Client

func init() {
	log.Println("Initializing Redis.")

	client, err := initRedis(os.Getenv("REDIS_URL"))
	if err != nil {
		panic(err)
	}
	RedisInstance = client
}

func initRedis(www string) (*redis.Client, error) {
	redisURL, _ := url.Parse(www)
	password := ""

	if(redisURL.User != nil) {
		password, _ = redisURL.User.Password()
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisURL.Host,
		Password: password,
		DB:       0,
		MaxRetries: 2,
	})

	err := client.Ping().Err()
	if err != nil {
		return nil, err
	}
	return client, nil
}

// TokenIsInBlacklist checks whether a token is blacklisted.
func TokenIsInBlacklist(tokenString string) bool {
	_, err := RedisInstance.Get(tokenString).Result()
	if err == redis.Nil {
		return false
	}
	return true
}

// TokenInvalidate invalidates a given token
func TokenInvalidate(token string, ttl time.Duration) error {
	return RedisInstance.Set(token, token, ttl).Err()
}

// AddActivity adds an activity to a user.
func AddActivity(userID uint, score float64, data string) error {
	return RedisInstance.ZAdd(fmt.Sprintf("%v", userID), redis.Z{Score: score, Member: data}).Err()
}

// GetActivities returns the activities of a user in the specified timeframe.
func GetActivities(userID uint, start int64, end int64) ([]string, error) {
	return RedisInstance.ZRevRangeByScore(fmt.Sprintf("%v", userID), redis.ZRangeByScore{Min: fmt.Sprintf("%v", start), Max: fmt.Sprintf("%v", end)}).Result()
}
