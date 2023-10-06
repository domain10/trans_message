package controllers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"trans_message/my_vendor/core"

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
 * 注册应用
 */
func (_ *App) Register(ctx *core.Context) {
	var data RegisterData
	if rawData, err := ctx.GetRawData(); err == nil {
		err = json.Unmarshal(rawData, &data)
		if err != nil || data.Name == "" || data.Query_url == "" || data.Notify_url == "" {
			ctx.Fail(http.StatusBadRequest, "Please pass the correct JSON data", nil)
			return
		} else if isValid, _ := regexp.MatchString(`^[\w-]+$`, data.Name); !isValid {
			ctx.Fail(http.StatusBadRequest, "The name must be composed of letters, numbers and underscores", nil)
			return
		} else if data.Query_url == data.Notify_url {
			ctx.Fail(http.StatusBadRequest, "Query address and notification address cannot be the same", nil)
			return
		} else if isValid, _ := regexp.MatchString(`^(http)s?://[\S]+$`, data.Query_url); !isValid {
			ctx.Fail(http.StatusBadRequest, "Invalid query url", nil)
		} else if isValid, _ := regexp.MatchString(`^(http)s?://[\S]+$`, data.Notify_url); !isValid {
			ctx.Fail(http.StatusBadRequest, "Invalid notify url", nil)
		}
		appkey := middleware.GenerateAppkey()
		//-------------db---------------
		model := new(models.Application)
		modelData := &models.Application{App_key: appkey, Ip_arr: ctx.ClientIP(), Name: data.Name, Query_url: data.Query_url, Notify_url: data.Notify_url, Describe: data.Describe}
		if err = model.RegisterOp(modelData); err == nil {
			ctx.Success(map[string]interface{}{"app_key": appkey})
		} else {
			ctx.Fail(http.StatusInternalServerError, err.Error(), nil)
		}
	} else {
		ctx.Fail(http.StatusBadRequest, "No parameters", nil)
	}
}
