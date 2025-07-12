// Path: ./service/redis_service/chitchat_cache.go

package redis_service

import (
	"dialogTree/global"
	"time"
)

func set(key string, value string, expire time.Duration) {
	global.Redis.Set(key, value, expire)
}

func get(key string) string {
	return global.Redis.Get(key).Val()
}
