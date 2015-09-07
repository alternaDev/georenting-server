package models

import (
	"fmt"
	"net/url"
	"os"

	"gopkg.in/redis.v3"
)

// RedisInstance is a usable redis instance.
var RedisInstance = initRedis(os.Getenv("REDIS_URL"))

func initRedis(www string) *redis.Client {
	redisURL, _ := url.Parse(www)
	password, _ := redisURL.User.Password()

	client := redis.NewClient(&redis.Options{
		Addr:     redisURL.Host,
		Password: password,
		DB:       0,
	})

	pong, err := client.Ping().Result()

	fmt.Println(pong, err)

	return client
}
