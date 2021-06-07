package cache

import (
	//	"errors"
	"fmt"
	"os"

	//"trans_message/middleware/log"
	"github.com/gomodule/redigo/redis"
)

var (
	comRedis   *redis.Pool
	db         int    = 0
	messageKey string = "stream:message_list"
	field      string = "data"
	//messageKeyWait string = "list:message_queue:wait"

)

type Handler struct {
	RedisCache
}

func init() {
	var err error
	if comRedis, err = GetCache("default", db); err != nil {
		fmt.Println("redis initialization error: ", err)
		os.Exit(1)
	}
}

func (self *Handler) On() *Handler {
	self.p = comRedis
	return self
}

// func ReplayToArray(reply interface{}, err error) (string, error) {
// 	if err != nil {
// 		return nil, err
// 	}
// 	switch reply := reply.(type) {
// 	case []interface{}:
// 		for _, arg := range reply {
// 			ReplayToArray(arg,err)
// 			var streamData []map[string][string]
// 			streamData = append(streamData)
// 		}
// 	case []byte:
// 		return string(reply), nil
// 	case string:
// 		return reply, nil
// 	case nil:
// 		return nil, errors.New("redigo: nil returned")
// 	case Error:
// 		return nil, reply
// 	}
// 	return nil, fmt.Errorf("redigo: unexpected type for Values, got type %T", reply)
// }

func (self *Handler) Hget(value ...interface{}) (string, error) {
	return redis.String(self.Do("hget", value...))
}

func (self *Handler) Hset(value ...interface{}) (int, error) {
	return redis.Int(self.Do("hset", value...))
}

func (self *Handler) Get(value ...interface{}) (string, error) {
	return redis.String(self.Do("get", value...))
}

func (self *Handler) Set(value ...interface{}) (string, error) {
	return redis.String(self.Do("set", value...))
}

func (self *Handler) Del(value ...interface{}) (int, error) {
	return redis.Int(self.Do("del", value...))
}

func (self *Handler) Expire(value ...interface{}) (int, error) {
	return redis.Int(self.Do("expire", value...))
}

/*
func (self *Handler) TTL(value ...interface{}) (int64, error) {
	return redis.Int64(self.Do("ttl", value...))
}

func (self *Handler) Xadd(data string) (string, error) {
	return redis.String(self.Do("xadd", messageKey, "*", field, data))
}

func (self *Handler) Xread(mid string) (result string, err error) {
	result = ""
	//ReplayToArray(self.Do("xrange", messageKey, mid, mid))
	if tmpVal, err := redis.MultiBulk(self.Do("xrange", messageKey, mid, mid)); err == nil {
		if tmpVal, err = redis.Values(tmpVal[0], err); err == nil {
			if tmpVal, err = redis.Values(tmpVal[1], err); err == nil {
				if tmpVal2, err := redis.String(tmpVal[1], err); err == nil {
					result = tmpVal2
				}
			}
		}
	}
	return
}
*/
