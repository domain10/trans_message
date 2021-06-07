package routers

import (
	"trans_message/my_vendor/core"

	"trans_message/controllers"
	"trans_message/middleware"

	"github.com/gin-gonic/gin"
)

func Register() *gin.Engine {
	r := gin.New()
	//
	//r.Use(gin.Logger())
	r.Use(core.Handle(middleware.HandleErrors))
	r.Use(core.Handle(middleware.Auth))

	r.NoRoute(core.Handle(middleware.Nonexistent))
	r.NoMethod(core.Handle(middleware.Nonexistent))
	//
	app := new(controllers.App)
	r.POST("/register", core.Handle(app.Register))
	//
	trans := new(controllers.Transaction)
	v1 := r.Group("/message")
	//v1.Use(core.Handle(middleware.Auth))
	{
		v1.POST("/create", core.Handle(trans.Create))
		v1.PUT("/confirm", core.Handle(trans.Confirm))
		v1.PUT("/callback", core.Handle(trans.NotifyCallback))
		//
		// v1.POST("/create", core.Handle(trans.CreateR))
		// v1.PUT("/submit/:mid", core.Handle(trans.SubmitR))
	}
	return r
}
