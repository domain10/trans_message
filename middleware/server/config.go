package server

import (
	"fmt"
	"os"

	"github.com/go-ini/ini"
)

// 环境配置文件，统一操作
type Env struct {
	DEBUG               bool
	SERVER_ID           int64
	LATEST_GENERATED_ID string
	HTTP_PORT           string

	ACCESS_LOG      string
	ERROR_LOG       string
	SERVICE_ERR_LOG string
	INFO_LOG        string

	Ding_news_url string
	//Ding_username string
	Ding_userid string
}

var env Env

func init() {
	var err error
	var cfg *ini.File
	// load配置
	cfg, err = ini.Load("conf/app.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	cfg.BlockMode = false
	mode := cfg.Section("").Key("app_mode").String()
	if "develop" == mode {
		env.DEBUG = true
	}
	env.ACCESS_LOG = cfg.Section("").Key("access_log").String()
	env.ERROR_LOG = cfg.Section("").Key("error_log").String()
	env.SERVICE_ERR_LOG = cfg.Section("").Key("service_err_log").String()
	env.INFO_LOG = cfg.Section("").Key("info_log").String()
	env.SERVER_ID, err = cfg.Section("").Key("server_id").Int64()
	if err != nil {
		fmt.Printf("Server_id configuration error: %v", err)
		os.Exit(1)
	}
	env.LATEST_GENERATED_ID = cfg.Section("").Key("latest_generated_id").String()
	env.HTTP_PORT = cfg.Section("").Key("http_port").String()
	//
	env.Ding_news_url = cfg.Section("").Key("ding_news_url").String()
	// env.Ding_username = cfg.Section("").Key("ding_username").String()
	env.Ding_userid = cfg.Section("").Key("ding_userid").String()
}

func GetConfig() *Env {
	return &env
}

func SetConfig(k, v string) {
	if cfg, err := ini.Load("conf/app.ini"); err == nil {
		cfg.Section("").Key(k).SetValue(v)
		if err = cfg.SaveTo("conf/app.ini"); err == nil {
			env.LATEST_GENERATED_ID = v
		}
	}

}
