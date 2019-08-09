package redis

import (
	"daisy/logger"
	"github.com/go-redis/redis"
)

func New(url string) *redis.Client {
	opt, err := redis.ParseURL(url)
	if err != nil {
		logger.Error("解析链接字符串", url, "失败:", err.Error())
	}
	client := redis.NewClient(opt)
	_, err = client.Ping().Result()
	if err != nil {
		logger.Error("连接", url, "失败:", err.Error())
	}
	logger.Info("连接到:", url)
	return client
}