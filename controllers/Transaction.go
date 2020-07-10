package controllers

import (
	"core"
	"encoding/json"
	"net/http"
	"time"

	"trans_message/middleware"
	"trans_message/middleware/log"
	"trans_message/models"
	"trans_message/models/cache"
	//"github.com/gin-gonic/gin"
)

type Transaction struct {
	Base
}

type RequestData struct {
	Mid          string `json:"mid"`
	App_key      string `json:"app_key"`
	From_address string `json:"from_address"`
	Message_type int    `json:"message_type"`
	List         []MessageData
	Status       int   `json:"status"`
	Notify_count int   `json:"notify_count"`
	Ctime        int64 `json:"ctime"`
	Utime        int64 `json:"utime"`
}

/**
 * @准备事务消息
 */
func (this *Transaction) Create(ctx *core.Context) {
	var data RequestData
	rawData, err := ctx.GetRawData()
	if err != nil {
		ctx.Fail(http.StatusBadRequest, "No parameters", nil)
		return
	}
	err = json.Unmarshal(rawData, &data)
	if err != nil || data.App_key == "" || data.Message_type <= 0 || data.List == nil {
		ctx.Fail(http.StatusBadRequest, "Please pass the correct JSON data", nil)
		return
	}
	content, _ := json.Marshal(data.List)
	//-------------check------------
	if !this.CheckAppExist(data.App_key) {
		ctx.Fail(http.StatusBadRequest, "Please complete the registration first", nil)
	} else if this.CheckDuplicated(data.App_key + string(content)) {
		ctx.Fail(http.StatusBadRequest, "Please do not submit duplicate messages", nil)
	} else {
		mid := middleware.GenerateId()
		//-------------db---------------
		model := new(models.MessageList)
		if id := model.Insert(mid, data.App_key, ctx.ClientIP(), data.Message_type, string(content)); id > 0 {
			ctx.Success(map[string]interface{}{"mid": mid})
		} else {
			ctx.Fail(http.StatusInternalServerError, "Failed to save message", nil)
		}
	}

}

/**
 * @提交事务
 */
func (this *Transaction) Submit(ctx *core.Context) {
	mid := ctx.Param("mid")
	if mid != "" {
		model := new(models.MessageList)
		data := model.GetDataByid(mid)
		if data.MessageType > 0 {
			if err := this.NotifyMessage(data); err == nil {
				ctx.Success(nil)
			} else {
				ctx.Fail(http.StatusInternalServerError, "Failure to submit,the message content format is incorrect", nil)
			}
		} else {
			ctx.Fail(http.StatusBadRequest, "No message with this mid="+mid, nil)
		}
	} else {
		ctx.Fail(http.StatusBadRequest, "No parameters", nil)
	}
}

/**
 * @准备事务消息R
 */
func (_ *Transaction) CreateR(ctx *core.Context) {
	var data RequestData
	if rawData, err := ctx.GetRawData(); err == nil {
		if err = json.Unmarshal(rawData, &data); err == nil && data.Message_type > 0 && data.List != nil {
			//
			data.From_address = ctx.ClientIP()
			data.Status = 1
			data.Ctime = time.Now().Unix()
			//-------------redis---------------
			obj := new(cache.Handler).On()
			content, _ := json.Marshal(data)
			if mid, err := obj.Xadd(string(content)); err == nil {
				ctx.Success(map[string]interface{}{"mid": mid})
			} else {
				ctx.Fail(http.StatusInternalServerError, "Failed to save message", nil)
			}
		} else {
			ctx.Fail(http.StatusBadRequest, "Please pass the correct JSON data", nil)
		}
	} else {
		ctx.Fail(http.StatusBadRequest, "No parameters", nil)
	}
}

/**
 * @提交事务R
 */
func (this *Transaction) SubmitR(ctx *core.Context) {
	mid := ctx.Param("mid")
	if mid != "" {
		var data RequestData
		//-------------redis---------------
		obj := new(cache.Handler).On()
		if res, err := obj.Xread(mid); err == nil {
			_ = json.Unmarshal([]byte(res), &data)
		} else {
			log.Error(err)
		}
		if data.Message_type > 0 {
			data.Mid = mid
			if err := this.NotifyMessageR(data); err == nil {
				ctx.Success(nil)
			} else {
				ctx.Fail(http.StatusInternalServerError, "Failure to submit,the message content format is incorrect", err)
			}
		} else {
			ctx.Fail(http.StatusBadRequest, "No message with this mid="+mid, nil)
		}
	} else {
		ctx.Fail(http.StatusBadRequest, "No parameters", nil)
	}
}
