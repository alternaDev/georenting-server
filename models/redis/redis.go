package redis

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"gopkg.in/redis.v3"
)

const (
	// BalanceNameExpenseRent describes the name for the Balance expense rent.
	BalanceNameExpenseRent = "expense-rent"
	// BalanceNameExpenseGeoFence describes the name for the Balance expense fence.
	BalanceNameExpenseGeoFence = "expense-fence"
	// BalanceNameEarningsRent describes the name for the Balance earning rent.
	BalanceNameEarningsRent = "earnings-rent"
	// OldestActivityAgeDays specifies the maximum time activities should be saved.
	// everything after this gets deleted.
	OldestActivityAgeDays = 30
)

type balanceRecord struct {
	Time  int64
	Value float64
}

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

	if redisURL.User != nil {
		password, _ = redisURL.User.Password()
	}

	client := redis.NewClient(&redis.Options{
		Addr:       redisURL.Host,
		Password:   password,
		DB:         0,
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
	id := fmt.Sprintf("%v", userID)

	oldest := time.Now().AddDate(0, 0, -OldestActivityAgeDays).Unix()
	err := RedisInstance.ZRemRangeByScore(id, "-inf", strconv.FormatInt(oldest, 10)).Err()
	if err != nil {
		return nil, err
	}

	return RedisInstance.ZRevRangeByScore(id, redis.ZRangeByScore{Min: strconv.FormatInt(start, 10), Max: strconv.FormatInt(end, 10)}).Result()
}

// GetBalanceRecordName returns the name of a BR for the user.
func GetBalanceRecordName(id uint, name string) string {
	return fmt.Sprintf("%d-%v", id, name)
}

// AddBalanceRecord adds a balance record to Redis.
func AddBalanceRecord(id string, value float64) error {
	now := time.Now().Unix()

	bytes, err := json.Marshal(balanceRecord{
		Time:  now,
		Value: value})

	if err != nil {
		return err
	}
	return RedisInstance.ZAdd(id, redis.Z{Score: float64(now), Member: string(bytes[:])}).Err()
}

// GetBalance returns the Value of the Balance Set
func GetBalance(id string) (float64, error) {
	sevenDaysAgo := time.Now().AddDate(0, 0, -7).Unix()

	err := RedisInstance.ZRemRangeByScore(id, "-inf", strconv.FormatInt(sevenDaysAgo, 10)).Err()
	if err != nil {
		return 0, err
	}

	r := RedisInstance.ZRangeByScore(id, redis.ZRangeByScore{Min: "-inf", Max: "+inf"})

	if r.Err() != nil {
		return 0, r.Err()
	}

	sum := 0.0

	for _, v := range r.Val() {
		var record balanceRecord
		err := json.Unmarshal([]byte(v), &record)
		if err != nil {
			return 0, err
		}
		sum += record.Value
	}

	return sum, nil
}
