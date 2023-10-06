package controllers

import (
	"encoding/json"
	"net/http"
	"trans_message/my_vendor/core"

	// "strconv"
	// "time"

	"trans_message/middleware"
	"trans_message/models"
	// "trans_message/middleware/log"
	// "trans_message/models/cache"
	//"github.com/gin-gonic/gin"
)

type Transaction struct {
	Base
}

type createData struct {
	Mid          int64  `json:"mid"`
	From_app     string `json:"from_app"`
	From_address string `json:"from_address"`
	Message_type int    `json:"message_type"`
	List         []MessageData
	Status       int   `json:"status"`
	Notify_count int   `json:"notify_count"`
	Ctime        int64 `json:"ctime"`
	Utime        int64 `json:"utime"`
}

type confirmData struct {
	Mid     int64  `json:"mid"`
	Operate string `json:"operate"`
	Msg     string `json:"msg"`
}

type notifyCallbackData struct {
	Mid      int64  `json:"mid"`
	From_app string `json:"from_app"`
	Status   int    `json:"status"`
	Msg      string `json:"msg"`
}

/**
 * 准备事务消息
 */
func (this *Transaction) Create(ctx *core.Context) {
	var data createData
	rawData, err := ctx.GetRawData()
	if err != nil {
		ctx.Fail(http.StatusBadRequest, "No parameters", nil)
		return
	}
	err = json.Unmarshal(rawData, &data)
	if err != nil || data.From_app == "" || data.Message_type <= 0 {
		ctx.Fail(http.StatusBadRequest, "Please pass the correct JSON data", nil)
		return
	} else if len(data.List) < 1 {
		ctx.Fail(http.StatusBadRequest, "Please pass the message content", nil)
		return
	}
	//content, _ := json.Marshal(data.List)

	//-------------check------------
	if !this.CheckAppExist(data.From_app) {
		ctx.Fail(http.StatusBadRequest, "Please complete the registration first", nil)
	} else {
		// 找出通知url
		for k, msg := range data.List {
			//
			if msg.To_app == "" || msg.Content == "" {
				ctx.Fail(http.StatusBadRequest, "the submitted field `list` data is incorrect", nil)
				return
			}

			if msg.To_url != "" {
				continue
			}
			// 获取应用信息
			toAppInfo := this.GetAppInfos(msg.To_app)
			if toAppInfo.Notify_url == "" {
				ctx.Fail(http.StatusBadRequest, "to_app("+msg.To_app+") url does not exist", nil)
				return
			}
			data.List[k].To_url = toAppInfo.Notify_url
		}
		// 要改为加入管道
		content, _ := json.Marshal(data.List)
		listStr := string(content)
		if this.CheckDuplicated(data.From_app + listStr) {
			ctx.Fail(http.StatusBadRequest, "Please do not submit duplicate messages", nil)
			return
		}

		if data.Mid == 0 {
			data.Mid = middleware.GenerateId()
		}
		// 入库
		model := new(models.MessageList)
		if id := model.Insert(data.Mid, data.From_app, ctx.ClientIP(), data.Message_type, listStr); id > 0 {
			msgC := middleware.GetMsgChannel()
			select {
			case msgC <- true:
			default:
			}
			ctx.Success(map[string]int64{"mid": data.Mid})
		} else {
			ctx.Fail(http.StatusInternalServerError, "Failed to save message:", nil)
		}
	}
}

/**
 * 提交事务
 */
func (this *Transaction) Confirm(ctx *core.Context) {
	//mid := ctx.Param("mid")
	var (
		data confirmData
	)
	rawData, err := ctx.GetRawData()
	if err != nil {
		ctx.Fail(http.StatusBadRequest, "No parameters", nil)
		return
	}
	err = json.Unmarshal(rawData, &data)
	if err != nil || data.Mid < 0 || data.Operate == "" {
		ctx.Fail(http.StatusBadRequest, "Please pass the correct JSON data", nil)
		return
	}

	model := new(models.MessageList)
	tmpData := model.GetDataBymid(data.Mid)
	if tmpData.ID <= 0 {
		ctx.Fail(http.StatusBadRequest, "No message", nil)
		return
	} else if tmpData.Status >= models.NOTIFYING_MSG {
		ctx.Fail(http.StatusBadRequest, "confirmed message", nil)
		return
	}

	switch data.Operate {
	case "submit":
		if _, err := this.NotifyMessage(tmpData); err == nil {
			ctx.Success(nil)
		} else {
			ctx.Fail(http.StatusInternalServerError, "Failure to submit", nil)
		}
	case "cancel":
		model.ModifyStatus(data.Mid, "", data.Msg, models.CANCEL_MSG)
		ctx.Success(nil)
	}
	this.DelCachedMsg(tmpData.AppName + tmpData.List)
}

/**
 * 执行事务通知回调
 */
func (this *Transaction) NotifyCallback(ctx *core.Context) {
	var (
		data notifyCallbackData
		list []MessageData
	)
	rawData, err := ctx.GetRawData()
	if err != nil {
		ctx.Fail(http.StatusBadRequest, "No parameters", nil)
		return
	}
	err = json.Unmarshal(rawData, &data)
	if err != nil || data.Mid < 0 || data.From_app == "" {
		ctx.Fail(http.StatusBadRequest, "Please pass the correct JSON data", nil)
		return
	}

	model := new(models.MessageList)
	tableData := model.GetDataBymid(data.Mid)
	if tableData.ID <= 0 {
		ctx.Fail(http.StatusBadRequest, "the mid for the callback does not exist", nil)
		return
	}

	if err = json.Unmarshal([]byte(tableData.List), &list); err == nil {
		notifyStatus := models.FAIL_MSG
		i := 0
		for k, v := range list {
			// 是该应用的回调
			if v.To_app == data.From_app && 1 == data.Status {
				// 成功
				list[k].Status = 1
				data.Msg = "success"
			}

			if 1 == list[k].Status {
				i++
			}
		}

		tmpContent, _ := json.Marshal(list)
		if len(list) == i {
			notifyStatus = models.SUCCESS_MSG
		}
		model.ModifyStatus(data.Mid, string(tmpContent), data.Msg, notifyStatus)
	}
	ctx.Success(nil)
}

/**
 * 准备事务消息R
 */
func (_ *Transaction) CreateR(ctx *core.Context) {
	// var data createData
	// if rawData, err := ctx.GetRawData(); err == nil {
	// 	if err = json.Unmarshal(rawData, &data); err == nil && data.Message_type > 0 && data.List != nil {
	// 		//
	// 		data.From_address = ctx.ClientIP()
	// 		data.Status = 1
	// 		data.Ctime = time.Now().Unix()
	// 		//-------------redis---------------
	// 		obj := new(cache.Handler).On()
	// 		content, _ := json.Marshal(data)
	// 		if mid, err := obj.Xadd(string(content)); err == nil {
	// 			ctx.Success(map[string]interface{}{"mid": mid})
	// 		} else {
	// 			ctx.Fail(http.StatusInternalServerError, "Failed to save message", nil)
	// 		}
	// 	} else {
	// 		ctx.Fail(http.StatusBadRequest, "Please pass the correct JSON data", nil)
	// 	}
	// } else {
	// 	ctx.Fail(http.StatusBadRequest, "No parameters", nil)
	// }
}

/**
 * 提交事务R
 */
func (this *Transaction) SubmitR(ctx *core.Context) {
	// mid := ctx.Param("mid")
	// if mid != "" {
	// 	var data createData
	// 	//-------------redis---------------
	// 	obj := new(cache.Handler).On()
	// 	if res, err := obj.Xread(mid); err == nil {
	// 		_ = json.Unmarshal([]byte(res), &data)
	// 	} else {
	// 		log.Error(err)
	// 	}
	// 	if data.Message_type > 0 {
	// 		data.Mid = mid
	// 		if err := this.NotifyMessageR(data); err == nil {
	// 			ctx.Success(nil)
	// 		} else {
	// 			ctx.Fail(http.StatusInternalServerError, "Failure to submit,the message content format is incorrect", err)
	// 		}
	// 	} else {
	// 		ctx.Fail(http.StatusBadRequest, "No message with this mid="+mid, nil)
	// 	}
	// } else {
	// 	ctx.Fail(http.StatusBadRequest, "No parameters", nil)
	// }
}
