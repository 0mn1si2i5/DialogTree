// Path: ./service/redis_service/chitchat_cache.go

package redis_service

import (
	"dialogTree/global"
	"fmt"
	"sort"
	"time"
)

func CacheChitChat(key, field, prompt, answer, summary string) {
	skey := fmt.Sprintf("cc_sum_%s", key)
	hpkey := fmt.Sprintf("cc_his_pmt_%s", key)
	hakey := fmt.Sprintf("cc_his_ans_%s", key)

	pByte := []byte(get(skey))
	if len(pByte) > 500 {
		pByte = pByte[:500]
	}

	newSummary := fmt.Sprintf("%s;%s", string(pByte), summary)

	set(skey, newSummary, 12*time.Hour)
	hset(hpkey, field, prompt)
	setExpire(hakey, 12*time.Hour)
	hset(hakey, field, answer)
	setExpire(hpkey, 12*time.Hour)

	hafields, hpfields := hgetFields(hakey), hgetFields(hpkey)
	if len(hafields) > 3 {
		sort.Strings(hafields)
		for i := range len(hafields) - 3 {
			hdel(hakey, hafields[i])
		}
	}
	if len(hpfields) > 3 {
		sort.Strings(hpfields)
		for i := range len(hpfields) - 3 {
			hdel(hpkey, hpfields[i])
		}
	}
}

func GetChitChat(key string) (prompts, answers map[string]string, summary string) {
	skey := fmt.Sprintf("cc_sum_%s", key)
	hpkey := fmt.Sprintf("cc_his_pmt_%s", key)
	hakey := fmt.Sprintf("cc_his_ans_%s", key)

	prompts = global.Redis.HGetAll(hpkey).Val()
	answers = global.Redis.HGetAll(hakey).Val()
	summary = get(skey)
	return
}

func DelChitChat(key string) {
	skey := fmt.Sprintf("cc_sum_%s", key)
	hpkey := fmt.Sprintf("cc_his_pmt_%s", key)
	hakey := fmt.Sprintf("cc_his_ans_%s", key)
	global.Redis.Del(skey)
	global.Redis.Del(hpkey)
	global.Redis.Del(hakey)
}
