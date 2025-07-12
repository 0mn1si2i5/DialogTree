// Path: ./service/redis_service/enter.go

package redis_service

import (
	"dialogTree/global"
	"time"
)

func set(key string, value string, expire time.Duration) {
	global.Redis.Set(key, value, expire)
}

func hset(key string, field string, value string) {
	global.Redis.HSet(key, field, value)
}

func get(key string) string {
	return global.Redis.Get(key).Val()
}

func hget(key string, field string) string {
	return global.Redis.HGet(key, field).Val()
}

func hgetFields(key string) []string {
	return global.Redis.HKeys(key).Val()
}

func setExpire(key string, expire time.Duration) {
	global.Redis.Expire(key, expire)
}

func hdel(key string, field string) {
	global.Redis.HDel(key, field)
}
