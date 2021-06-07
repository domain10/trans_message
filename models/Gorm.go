package models

import (
	"fmt"
	"os"
	"time"
	"trans_message/middleware/server"

	"github.com/go-ini/ini"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var ormArr []*gorm.DB

func init() {
	var err error
	if ormArr, err = GetConnect("default"); err != nil {
		fmt.Printf("mysql initialization error: %v", err)
		os.Exit(1)
	}
}

func GetOrm(mid int64, connArr []*gorm.DB) *gorm.DB {
	if connArr == nil {
		connArr = ormArr
	}
	mod := mid % int64(len(connArr))
	return connArr[mod]
}

func GetConnect(name string) ([]*gorm.DB, error) {
	var connArr []*gorm.DB
	var err error
	var cfg *ini.File
	var deploy int
	var maxIdleConns int
	var maxOpenConns int
	// load配置
	cfg, err = ini.Load(server.RunPath() + "/conf/database.ini")
	if err != nil {
		fmt.Printf("%v", err)
		os.Exit(1)
	}
	// 数据库名称
	dbname := cfg.Section(name).Key("dbname").String()
	// 最大空闲连接数
	maxIdleConns = cfg.Section(name).Key("max_idle_conns").MustInt(10)
	// 最大打开的连接数
	maxOpenConns = cfg.Section(name).Key("max_open_conns").MustInt(30)
	//部署方式
	deploy = cfg.Section(name).Key("deploy").MustInt(1)
	switch deploy {
	case 1:
		// 主机
		host := cfg.Section(name).Key("host").String()
		// 端口
		port := cfg.Section(name).Key("port").String()
		// 用户名
		username := cfg.Section(name).Key("username").String()
		// 密码
		password := cfg.Section(name).Key("password").String()

		dsn := username + ":" + password + "@tcp(" + host + ":" + port + ")/" + dbname + "?charset=utf8&parseTime=true&loc=Local"
		conn, err := gorm.Open("mysql", dsn)
		if err != nil {
			return nil, err
		}
		conn.DB().SetMaxIdleConns(maxIdleConns)
		conn.DB().SetMaxOpenConns(maxOpenConns)
		conn.DB().SetConnMaxLifetime(time.Hour)
		conn.SingularTable(true)
		connArr = append(connArr, conn)
	case 2:
		fallthrough
	default:
		// 主机
		hostArr := cfg.Section(name).Key("host").Strings(",")
		// 端口
		portArr := cfg.Section(name).Key("port").Strings(",")
		// 用户名
		usernameArr := cfg.Section(name).Key("username").Strings(",")
		// 密码
		passwordArr := cfg.Section(name).Key("password").Strings(",")
		length := len(hostArr)
		port := portArr[0]
		username := portArr[0]
		password := portArr[0]
		for k, host := range hostArr {
			if len(portArr) == length {
				port = portArr[k]
			}
			if len(usernameArr) == length {
				username = usernameArr[k]
			}
			if len(passwordArr) == length {
				password = passwordArr[k]
			}
			dsn := username + ":" + password + "@tcp(" + host + ":" + port + ")/" + dbname + "?charset=utf8&parseTime=true&loc=Local"
			conn, err := gorm.Open("mysql", dsn)
			if err != nil {
				return nil, err
			}
			conn.DB().SetMaxIdleConns(maxIdleConns)
			conn.DB().SetMaxOpenConns(maxOpenConns)
			conn.DB().SetConnMaxLifetime(time.Minute * 10)
			conn.SingularTable(true)
			connArr = append(connArr, conn)
		}
	}
	return connArr, nil
}
