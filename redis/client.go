package redis

import (
	"fmt"
	"github.com/go-redis/redis"
)

func New(url string) *redis.Client {
	opt, err := redis.ParseURL(url)
	if err != nil {
		fmt.Println("解析链接字符串", url, "失败:", err.Error())
	}
	client := redis.NewClient(opt)
	err = client.Ping().Err()
	if err != nil {
		fmt.Println("连接", url, "失败:", err.Error())
	}
	fmt.Println("连接到:", url)
	return client
}
