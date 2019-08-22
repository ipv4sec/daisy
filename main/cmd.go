package main

import (
	"daisy/logger"
	"daisy/redis"
	"flag"
	"time"
)

var (
	sourceURL string
	targetURL string
	countNum int64
	matchStr string
)

func init() {
	//flag.StringVar(&sourceURL, "s", "redis://localhost:6379/0", "Source")
	//flag.StringVar(&targetURL, "t", "redis://user:passwd@localhost:6379/0", "Target")

	flag.StringVar(&sourceURL, "s", "redis://localhost:6379/0", "来源数据库")
	flag.StringVar(&targetURL, "t", "redis://localhost:6379/3", "目的数据库")
	flag.Int64Var(&countNum, "n", 1000, "每次扫描数量")
	flag.StringVar(&matchStr, "m", "*", "迁移匹配格式")
	// Redis TTL的值 和 Expire 的秒数不是同一个东西
}
func main() {
	flag.Parse()

	source := redis.New(sourceURL)
	target := redis.New(targetURL)

	var cursor uint64
	var err error

	cursor = 0

	for {
		var result []string
		result, cursor, err = source.Scan(cursor, matchStr, countNum).Result() // Redis >= 2.8
		if err != nil {
			logger.Error("遍历键名错误:", err)
		}
		logger.Info("此次迁移数量:", len(result))
		for i := 0; i < len(result); i++ {
			t, err := source.Type(result[i]).Result()
			if err != nil {
				logger.Error("获取键名为", result[i], "的类型错误:", err, )
				continue
			}
			switch t {
			case redis.TYPE_STRING:
				expireAt, err := source.TTL(result[i]).Result()
				if err != nil {
					logger.Error("获取", result[i], "的过期时间错误", err.Error())
					break
				}
				if expireAt == -2  * time.Second {
					// Redis >= 2.8
					break
				}
				value, err := source.Get(result[i]).Result()
				if err != nil {
					logger.Error("获取STRING键", result[i], "错误:", err.Error())
					break
				}
				if expireAt >= 0 {
					_, err = target.Set(result[i], value, expireAt).Result()
					if err != nil {
						logger.Error("保存", result[i], "错误:", err.Error())
					}
				}
			case redis.TYPE_SET:
				expireAt, err := source.TTL(result[i]).Result()
				if err != nil {
					logger.Error("获取", result[i], "的过期时间错误", err.Error())
					break
				}
				if expireAt == -2  * time.Second {
					break
				}
				value, err := source.SMembers(result[i]).Result()
				if err != nil {
					logger.Error("获取SET键", result[i], "错误:", err.Error())
					break
				}
				_, err = target.SAdd(result[i], value).Result()
				if err != nil {
					logger.Error("保存", result[i], "失败:", err.Error())
					break
				}
				if expireAt >= 0 {
					ok, err := target.Expire(result[i], expireAt).Result()
					if err != nil {
						logger.Error("设置", result[i], "过期时间错误:", err.Error())
						break
					}
					if !ok {
						logger.Error("设置", result[i], "过期时间失败")
					}
				}
			case redis.TYPE_ZSET:
				expireAt, err := source.TTL(result[i]).Result()
				if err != nil {
					logger.Error("获取", result[i], "的过期时间错误", err.Error())
					break
				}
				if expireAt == -2  * time.Second  {
					break
				}
				value, err := source.ZRangeWithScores(result[i], 0, -1).Result()
				if err != nil {
					logger.Error("获取ZSET键", result[i], "失败:", err.Error())
					break
				}
				_, err = target.ZAdd(result[i], value...).Result()
				if err != nil {
					logger.Error("保存", result[i], "失败:", err.Error())
					break
				}
				if expireAt >= 0 {
					ok, err := target.Expire(result[i], expireAt).Result()
					if err != nil {
						logger.Error("设置", result[i], "过期时间错误:", err.Error())
						break
					}
					if !ok {
						logger.Error("设置", result[i], "过期时间失败")
					}
				}
			case redis.TYPE_HASH:
				expireAt, err := source.TTL(result[i]).Result()
				if err != nil {
					logger.Error("获取", result[i], "的过期时间错误", err.Error())
					break
				}
				if expireAt == -2  * time.Second {
					break
				}
				value, err := source.HGetAll(result[i]).Result()
				if err != nil {
					logger.Error("获取HASH键", result[i], "失败:", err.Error())
					break
				}
				val := map[string]interface{}{}
				for k, v := range value {
					val[k] = v
				}
				_, err = target.HMSet(result[i], val).Result()
				if err != nil {
					logger.Error("保存", result[i], "失败:", err.Error())
					break
				}
				if expireAt >= 0 {
					ok, err := target.Expire(result[i], expireAt).Result()
					if err != nil {
						logger.Error("设置", result[i], "过期时间错误:", err.Error())
						break
					}
					if !ok {
						logger.Error("设置", result[i], "过期时间失败")
					}
				}
			case redis.TYPE_LIST:
				expireAt, err := source.TTL(result[i]).Result()
				if err != nil {
					logger.Error("获取", result[i], "的过期时间错误", err.Error())
					break
				}
				if expireAt == -2 * time.Second {
					break
				}
				value, err := source.LRange(result[i], 0, -1).Result()
				if err != nil {
					logger.Error("获取LIST键", result[i], "失败:", err.Error())
					break
				}
				for i, j := 0, len(value)-1; i < j; i, j = i+1, j-1 {
					value[i], value[j] = value[j], value[i]
				}
				_, err = target.LPush(result[i], value).Result()
				if err != nil {
					logger.Error("保存", result[i], "失败:", err.Error())
					break
				}
				if expireAt >= 0 {
					ok, err := target.Expire(result[i], expireAt).Result()
					if err != nil {
						logger.Error("设置", result[i], "过期时间错误:", err.Error())
						break
					}
					if !ok {
						logger.Error("设置", result[i], "过期时间失败")
					}
				}
			}
		}

		if cursor == 0 {
			logger.Info("迁移结束")
			break
		}
	}
}
