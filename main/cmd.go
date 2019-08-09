package main

import (
	"daisy/logger"
	"daisy/redis"
	"flag"
)

var (
	sourceURL string
	targetURL string
)

func init() {
	//flag.StringVar(&sourceURL, "s", "redis://localhost:6379/0", "Source")
	//flag.StringVar(&targetURL, "t", "redis://user:passwd@localhost:6379/0", "Target")

	flag.StringVar(&sourceURL, "s", "redis://localhost:6379/0", "Source")
	flag.StringVar(&targetURL, "t", "redis://localhost:6379/3", "Target")
}
func main() {
	flag.Parse()

	sourceClient := redis.New(sourceURL)
	tagetClient := redis.New(targetURL)

	var cursor uint64
	var err error

	cursor = 0

	for {
		var result []string
		result, cursor, err = sourceClient.Scan(cursor, "*", 1000).Result()
		if err != nil {
			logger.Error("遍历键名失败:", err)
		}
		logger.Info("此次迁移数量:", len(result))
		for i := 0; i < len(result); i++ {
			t, err := sourceClient.Type(result[i]).Result()
			if err != nil {
				logger.Error("获取键名为", result[i], "的类型错误:", err, )
				continue
			}
			switch t {
			case redis.TYPE_STRING:
				value, err := sourceClient.Get(result[i]).Result()
				if err != nil {
					logger.Error("获取STRING键", result[i], "失败:", err.Error())
				}
				sourceClient.PTTL()
				_, err = tagetClient.Set(result[i], value, -1).Result()
				if err != nil {
					logger.Error("保存", result[i], "失败:", err.Error())
				}
			case redis.TYPE_SET:
				value, err := sourceClient.SMembers(result[i]).Result()
				if err != nil {
					logger.Error("获取SET键", result[i], "失败:", err.Error())
				}
				_, err = tagetClient.SAdd(result[i], value).Result()
				if err != nil {
					logger.Error("保存", result[i], "失败:", err.Error())
				}
			case redis.TYPE_ZSET:
				value, err := sourceClient.ZRangeWithScores(result[i], 0, -1).Result()
				if err != nil {
					logger.Error("获取ZSET键", result[i], "失败:", err.Error())
				}
				_, err = tagetClient.ZAdd(result[i], value...).Result()
				if err != nil {
					logger.Error("保存", result[i], "失败:", err.Error())
				}
			case redis.TYPE_HASH:
				value, err := sourceClient.HGetAll(result[i]).Result()
				if err != nil {
					logger.Error("获取HASH键", result[i], "失败:", err.Error())
				}
				val := map[string]interface{}{}
				for k, v := range value {
					val[k] = v
				}
				_, err = tagetClient.HMSet(result[i], val).Result()
				if err != nil {
					logger.Error("保存", result[i], "失败:", err.Error())
				}
			case redis.TYPE_LIST:
				value, err := sourceClient.LRange(result[i], 0, -1).Result()
				if err != nil {
					logger.Error("获取LIST键", result[i], "失败:", err.Error())
				}
				_, err = tagetClient.LPush(result[i], value).Result()
				if err != nil {
					logger.Error("保存", result[i], "失败:", err.Error())
				}
			}
		}

		if cursor == 0 {
			break
		}
	}


}
