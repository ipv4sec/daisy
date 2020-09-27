package main

import (
	"daisy/config"
	"daisy/redis"
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"
	"time"
)

var (
	conf string
	c config.Config
)

func init() {
	flag.StringVar(&conf, "conf", "example.yaml", "配置文件")
}
func main() {
	flag.Parse()
	bytes, err := ioutil.ReadFile(conf)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(bytes, &c)
	if err != nil {
		panic(err)
	}
	sourceURL := fmt.Sprintf(
		"redis://:%s@%s/%v", c.Source.Auth, strings.Join(c.Source.Host, ","), c.Source.Database)
	targetURL := fmt.Sprintf(
		"redis://:%s@%s/%v", c.Source.Auth, strings.Join(c.Source.Host, ","), c.Source.Database)
	source := redis.New(sourceURL)
	target := redis.New(targetURL)

	cursor := uint64(0)
	startAt := time.Now()
	for {
		var result []string
		result, cursor, err = source.Scan(cursor, c.Source.Prefix + "*", 1000).Result() // Redis >= 2.8
		if err != nil {
			fmt.Println("遍历键名错误:", err)
		}
		if len(result) != 0 {
			fmt.Println("此次迁移数量:", len(result))
		}
		for i := 0; i < len(result); i++ {
			t, err := source.Type(result[i]).Result()
			if err != nil {
				fmt.Println("获取键名为", result[i], "的类型错误:", err, )
				continue
			}
			switch t {
			case redis.TYPE_STRING:
				expireAt, err := source.TTL(result[i]).Result()
				if err != nil {
					fmt.Println("获取", result[i], "的过期时间错误", err.Error())
					break
				}
				if expireAt == -2 {
					break
				}
				value, err := source.Get(result[i]).Result()
				if err != nil {
					fmt.Println("获取STRING键", result[i], "错误:", err.Error())
					break
				}
				newKeyName := strings.Replace(result[i], c.Source.Prefix, c.Target.Prefix, 1)
				err = target.Set(newKeyName, value, expireAt).Err()
				if err != nil {
					fmt.Println("保存", newKeyName, "错误:", err.Error())
				}
			case redis.TYPE_SET:
				expireAt, err := source.TTL(result[i]).Result()
				if err != nil {
					fmt.Println("获取", result[i], "的过期时间错误", err.Error())
					break
				}
				if expireAt == -2 {
					break
				}
				value, err := source.SMembers(result[i]).Result()
				if err != nil {
					fmt.Println("获取SET键", result[i], "错误:", err.Error())
					break
				}
				newKeyName := strings.Replace(result[i], c.Source.Prefix, c.Target.Prefix,1)
				err = target.SAdd(newKeyName, value).Err()
				if err != nil {
					fmt.Println("保存", newKeyName, "失败:", err.Error())
					break
				}
				if expireAt >= 0 {
					ok, err := target.Expire(newKeyName, expireAt).Result()
					if err != nil {
						fmt.Println("设置", newKeyName, "过期时间错误:", err.Error())
						break
					}
					if !ok {
						fmt.Println("设置", newKeyName, "过期时间失败")
					}
				}
			case redis.TYPE_ZSET:
				expireAt, err := source.TTL(result[i]).Result()
				if err != nil {
					fmt.Println("获取", result[i], "的过期时间错误", err.Error())
					break
				}
				if expireAt == -2  {
					break
				}
				value, err := source.ZRangeWithScores(result[i], 0, -1).Result()
				if err != nil {
					fmt.Println("获取ZSET键", result[i], "失败:", err.Error())
					break
				}
				newKeyName := strings.Replace(result[i], c.Source.Prefix, c.Target.Prefix, 1)
				err = target.ZAdd(newKeyName, value...).Err()
				if err != nil {
					fmt.Println("保存", newKeyName, "失败:", err.Error())
					break
				}
				if expireAt >= 0 {
					ok, err := target.Expire(newKeyName, expireAt).Result()
					if err != nil {
						fmt.Println("设置", newKeyName, "过期时间错误:", err.Error())
						break
					}
					if !ok {
						fmt.Println("设置", newKeyName, "过期时间失败")
					}
				}
			case redis.TYPE_HASH:
				expireAt, err := source.TTL(result[i]).Result()
				if err != nil {
					fmt.Println("获取", result[i], "的过期时间错误", err.Error())
					break
				}
				if expireAt == -2 {
					break
				}
				value, err := source.HGetAll(result[i]).Result()
				if err != nil {
					fmt.Println("获取HASH键", result[i], "失败:", err.Error())
					break
				}
				val := map[string]interface{}{}
				for k, v := range value {
					val[k] = v
				}
				newKeyName := strings.Replace(result[i], c.Source.Prefix, c.Target.Prefix, 1)
				err = target.HMSet(newKeyName, val).Err()
				if err != nil {
					fmt.Println("保存", newKeyName, "失败:", err.Error())
					break
				}
				if expireAt >= 0 {
					ok, err := target.Expire(newKeyName, expireAt).Result()
					if err != nil {
						fmt.Println("设置", newKeyName, "过期时间错误:", err.Error())
						break
					}
					if !ok {
						fmt.Println("设置", newKeyName, "过期时间失败")
					}
				}
			case redis.TYPE_LIST:
				expireAt, err := source.TTL(result[i]).Result()
				if err != nil {
					fmt.Println("获取", result[i], "的过期时间错误", err.Error())
					break
				}
				if expireAt == -2 {
					break
				}
				value, err := source.LRange(result[i], 0, -1).Result()
				if err != nil {
					fmt.Println("获取LIST键", result[i], "失败:", err.Error())
					break
				}
				for i, j := 0, len(value)-1; i < j; i, j = i+1, j-1 {
					value[i], value[j] = value[j], value[i]
				}
				newKeyName := strings.Replace(result[i], c.Source.Prefix, c.Target.Prefix, 1)
				err = target.LPush(newKeyName, value).Err()
				if err != nil {
					fmt.Println("保存", newKeyName, "失败:", err.Error())
					break
				}
				if expireAt >= 0 {
					ok, err := target.Expire(newKeyName, expireAt).Result()
					if err != nil {
						fmt.Println("设置", newKeyName, "过期时间错误:", err.Error())
						break
					}
					if !ok {
						fmt.Println("设置", newKeyName, "过期时间失败")
					}
				}
			}
		}

		if cursor == 0 {
			fmt.Println(fmt.Sprintf("迁移完成, 用时 %v", time.Since(startAt)))
			break
		}
	}
}
