package server

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-ini/ini"
)

// 环境配置文件，统一操作
type Env struct {
	DEBUG               bool
	SERVER_ID           int64
	LATEST_GENERATED_ID string
	HTTP_PORT           string

	INFO_LOG         string
	ACCESS_LOG       string
	ERROR_LOG        string
	Trans_system_err string
	Max_files        int

	Ding_news_key  string
	Ding_group_url string
	Alarm_interval string

	Default_interval string
	Con_notices      int
}

var (
	env      Env
	app_path string
)

func init() {
	var err error
	var cfg *ini.File
	// load配置
	cfg, err = ini.Load(RunPath() + "/conf/app.ini")
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	cfg.BlockMode = false
	mode := cfg.Section("").Key("app_mode").String()
	if "develop" == mode {
		env.DEBUG = true
	}

	env.INFO_LOG = cfg.Section("").Key("info_log").String()
	env.ACCESS_LOG = cfg.Section("").Key("access_log").String()
	env.ERROR_LOG = cfg.Section("").Key("error_log").String()
	env.Trans_system_err = cfg.Section("").Key("trans_system_err").String()
	// 最大日志文件数目
	env.Max_files, _ = cfg.Section("").Key("max_files").Int()
	if env.Max_files <= 0 {
		env.Max_files = 10
	}

	env.SERVER_ID, err = cfg.Section("").Key("server_id").Int64()
	if err != nil {
		fmt.Printf("Server_id configuration error: %v", err)
		os.Exit(1)
	}
	env.LATEST_GENERATED_ID = cfg.Section("").Key("latest_generated_id").String()
	env.HTTP_PORT = cfg.Section("").Key("http_port").String()
	//
	env.Ding_news_key = cfg.Section("").Key("ding_news_key").String()
	env.Ding_group_url = cfg.Section("").Key("ding_group_url").String()
	env.Alarm_interval = cfg.Section("").Key("alarm_interval").String()

	env.Default_interval = cfg.Section("").Key("default_interval").String()
	env.Con_notices, _ = cfg.Section("").Key("concurrent_notices").Int()
}

func GetConfig() *Env {
	return &env
}

func SetConfig(k, v string) {
	if cfg, err := ini.Load(RunPath() + "/conf/app.ini"); err == nil {
		cfg.Section("").Key(k).SetValue(v)
		if err = cfg.SaveTo(RunPath() + "/conf/app.ini"); err == nil {
			env.LATEST_GENERATED_ID = v
		}
	}

}

// 返回当前的运行目录
func RunPath() string {
	if app_path == "" {
		var err error
		app_path, err = filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			panic(err)
		}
	}
	return app_path
}
