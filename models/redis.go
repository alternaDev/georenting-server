package models

import (
	"fmt"
	"net/url"
	"os"

	"gopkg.in/redis.v3"
)

var RedisInstance = initRedis(os.Getenv("REDIS_URL"))

func initRedis(www string) *redis.Client {
	redisUrl, _ := url.Parse(www)
	password, _ := redisUrl.User.Password()
	client := redis.NewClient(&redis.Options{
		Addr:     redisUrl.Host,
		Password: password,
		DB:       0,
	})

	pong, err := client.Ping().Result()

	fmt.Println(pong, err)

	return client
}
