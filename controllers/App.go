package controllers

import (
	"core"
	"encoding/json"
	"net/http"

	"trans_message/middleware"
	"trans_message/models"
	//"github.com/gin-gonic/gin"
)

type App struct {
	Base
}

type RegisterData struct {
	Name       string `json:"name"`
	Query_url  string `json:"query_url"`
	Notify_url string `json:"notify_url"`
	Describe   string `json:"describe"`
	// App_key    string `json:"app_key"`
	// Ip         string `json:"ip"`
}

/**
 * @注册应用
 */
func (_ *App) Register(ctx *core.Context) {
	var data RegisterData
	if rawData, err := ctx.GetRawData(); err == nil {
		err = json.Unmarshal(rawData, &data)
		if err == nil && data.Name != "" && data.Query_url != "" && data.Notify_url != "" {
			appkey := middleware.GenerateAppkey()
			//-------------db---------------
			model := new(models.Application)
			modelData := &models.Application{App_key: appkey, Ip: ctx.ClientIP(), Name: data.Name, Query_url: data.Query_url, Notify_url: data.Notify_url, Describe: data.Describe}
			if err = model.RegisterOp(modelData); err == nil {
				ctx.Success(map[string]interface{}{"app_key": appkey})
			} else {
				ctx.Fail(http.StatusInternalServerError, err.Error(), nil)
			}
		} else {
			ctx.Fail(http.StatusBadRequest, "Please pass the correct JSON data", nil)
		}
	} else {
		ctx.Fail(http.StatusBadRequest, "No parameters", nil)
	}
}
