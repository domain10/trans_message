package cache

import (
	//	"errors"
	"fmt"
	"os"

	//"trans_message/middleware/log"
	"github.com/gomodule/redigo/redis"
)

type External struct {
	RedisCache
}

func init() {
	var err error
	if comRedis, err = GetCache("swooleQueue", 1); err != nil {
		fmt.Println("redis initialization error: ", err)
		os.Exit(1)
	}
}

func (self *External) On() *External {
	self.p = comRedis
	return self
}

func (self *External) DingNews(data string) (int, error) {
	c := self.p.Get()
	defer c.Close()
	script := `
	local qList = 'queue:'.. KEYS[1]
    local qHash = 'hash:queuer|'.. KEYS[1]
    redis.pcall('hincrby', qHash, KEYS[2], 1)
    return redis.pcall('lpush', qList, KEYS[2])
    `
	var getScript = redis.NewScript(2, script)
	return redis.Int(getScript.Do(c, "app\\internalletter\\queue\\SendDingQueue", data))
}
