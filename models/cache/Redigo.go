package cache

import (
	"errors"
	// "fmt"
	// "os"
	"strings"
	"time"
	"trans_message/middleware/server"

	"github.com/go-ini/ini"
	"github.com/gomodule/redigo/redis"
)

//var comRedis *redis.Pool

type RedisCache struct {
	p *redis.Pool
}

// func init() {
// 	var err error
// 	if comRedis, err = GetCache("default"); err != nil {
// 		fmt.Printf("%v", err)
// 		os.Exit(1)
// 	}
// }

func GetCache(name string, dbNum int) (redisPool *redis.Pool, err error) {
	var cfg *ini.File
	var maxIdleConns int
	var maxOpenConns int

	// load配置
	cfg, err = ini.Load(server.RunPath() + "/conf/redis.ini")
	if err != nil {
		return nil, err
	}
	// 主机
	host := cfg.Section(name).Key("host").String()
	// 端口
	port := cfg.Section(name).Key("port").String()
	// 密码
	password := cfg.Section(name).Key("password").String()
	//DB库
	if dbNum <= 0 && cfg.Section(name).HasKey("db") {
		if dbNum, err = cfg.Section(name).Key("db").Int(); err != nil {
			return nil, err
		}
	}
	// 最大空闲连接数
	if maxIdleConns, err = cfg.Section(name).Key("max_idle_conns").Int(); err != nil {
		return nil, err
	}
	// 最大打开的连接数
	if maxOpenConns, err = cfg.Section(name).Key("max_open_conns").Int(); err != nil {
		return nil, err
	}
	redisPool = &redis.Pool{
		MaxIdle:     maxIdleConns,
		MaxActive:   maxOpenConns,
		IdleTimeout: 240 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host+":"+port)
			//c, err := redis.DialTimeout("tcp", host+":"+port, 3*time.Second, 15*time.Second, 15*time.Second)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			if dbNum > 0 {
				if _, err := c.Do("SELECT", dbNum); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, nil
		},
	}
	c := redisPool.Get()
	defer c.Close()
	err = c.Err()
	if err != nil {
		redisPool = nil
	}
	return
}

func (rc *RedisCache) Do(command string, args ...interface{}) (reply interface{}, err error) {
	if strings.ToLower(command) == "select" {
		return nil, errors.New("Forbid switching of database in operation")
	}
	c := rc.p.Get()
	defer c.Close()
	return c.Do(command, args...)
}
