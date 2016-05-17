package redis

import (
	"net/url"
	"os"
	"fmt"
	"time"

	"gopkg.in/redis.v3"
)

// RedisInstance is a usable redis instance.
var RedisInstance = initRedis(os.Getenv("REDIS_URL"))

func initRedis(www string) *redis.Client {
	redisURL, _ := url.Parse(www)
	password := ""

	if(redisURL.User != nil) {
		password, _ = redisURL.User.Password()
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisURL.Host,
		Password: password,
		DB:       0,
	})

	return client
}

func TokenIsInBlacklist(tokenString string) bool {
	_, err := RedisInstance.Get(tokenString).Result()
	if err == redis.Nil {
		return false
	}
	return true
}

// InvalidateToken invalidates a given token
func TokenInvalidate(token string, ttl time.Duration) error {
	return RedisInstance.Set(token, token, ttl).Err()
}

func AddActivity(userID uint, score float64, data string) error {
	return RedisInstance.ZAdd(fmt.Sprintf("%v", userID), redis.Z{Score: score, Member: data}).Err()
}

func GetActivities(userID uint, start int64, end int64) ([]string, error) {
	return RedisInstance.ZRevRangeByScore(fmt.Sprintf("%v", userID), redis.ZRangeByScore{Min: fmt.Sprintf("%v", start), Max: fmt.Sprintf("%v", end)}).Result()
}
