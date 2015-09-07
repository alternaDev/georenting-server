package models

import (
	"fmt"
	"os"

	"gopkg.in/redis.v3"
)

var RedisInstance = initRedis(os.Getenv("REDIS_URL"))

func initRedis(url string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: url,
		DB:   0,
	})

	pong, err := client.Ping().Result()

	fmt.Println(pong, err)

	return client
}
