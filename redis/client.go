package redis

import "github.com/go-redis/redis"

func InitByUrl(url string) *redis.Client {
	opt, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}
	client := redis.NewClient(opt)
	_, err = client.Ping().Result()
	if err != nil {
		panic(err)
	}
	return client
}
