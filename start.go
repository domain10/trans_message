package main

import (
	"runtime"

	_ "trans_message/middleware/log"
	"trans_message/middleware/server"
	"trans_message/routers"

	"github.com/gin-gonic/gin"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	//-----------定时通知------------
	go scan()

	// 运行模式
	if server.GetConfig().DEBUG {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 注册路由
	r := routers.Register()

	// 加载模板文件
	//r.LoadHTMLGlob("templates/**/*")

	// 加载静态文件
	//r.Static("/static", "static")

	//r.Run(http_port)
	server.Run(r)
}
