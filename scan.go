package main

import (
	"encoding/json"
	"net/url"
	"strconv"
	"time"
	"trans_message/controllers"
	"trans_message/middleware"
	"trans_message/middleware/log"
	"trans_message/middleware/server"
	"trans_message/models"
	"trans_message/my_vendor/utils"
)

/**
 * 通知人为处理
 */
func notifyHuman(content string) (err error) {
	type DingMsg struct {
		Msgtype  string            `json:"msgtype"`
		Markdown map[string]string `json:"markdown"`
		At       map[string]bool   `json:"at"`
	}
	type DingResponse struct {
		Errcode int    `json:"errcode"`
		Errmsg  string `json:"errmsg"`
	}
	var data DingMsg
	var respData DingResponse
	data.Markdown = make(map[string]string)
	data.At = make(map[string]bool)

	data.Msgtype = "markdown"
	data.Markdown["title"] = "from " + server.GetConfig().Ding_news_key
	data.Markdown["text"] = "#### New alert from " + server.GetConfig().Ding_news_key + "\n"
	data.Markdown["text"] += content

	data.At["isAtAll"] = true
	pData, _ := json.Marshal(data)
	resp, err := middleware.RawRequest(server.GetConfig().Ding_group_url, "POST", string(pData), "")
	if err != nil {
		log.Error(server.GetConfig().Ding_group_url + ",发送(" + string(pData) + "),error:" + err.Error())
	} else if _ = json.Unmarshal(resp, &respData); respData.Errcode != 0 {
		log.Error(server.GetConfig().Ding_group_url + ",发送(" + string(pData) + "),response:" + string(resp))
	}

	return
}

func RunQueryNotify(data models.MessageList, appInfo models.Application, runNotifyCh <-chan bool) {
	defer func() {
		<-runNotifyCh
		// if err := recover(); err != nil {
		// 	log.ErrorStrace(err)
		// }
	}()
	model := new(models.MessageList)
	baseObj := new(controllers.Base)
	content := "事务消息通知失败，"
	needAlarm := false

	switch data.Status {
	case models.INIT_MSG:
		// 查询确认消息
		mid := strconv.FormatInt(data.Mid, 10)
		tmpTime := strconv.FormatInt(time.Now().Unix(), 10)
		sign := mid + "|" + tmpTime + "|" + appInfo.App_key
		sign = url.QueryEscape(utils.HmacSha256(sign, appInfo.App_key, true))
		resp, err := middleware.RequestOp(appInfo.Query_url, "GET", "mid="+mid+"&time="+tmpTime+"&sign="+sign, "")

		if err == nil {
			if "submit" == resp.Result {
				baseObj.NotifyMessage(data)
				return
			} else if "cancel" == resp.Result {
				model.ModifyStatus(data.Mid, "", resp.Msg, models.CANCEL_MSG)
				return
			}
		}
		upStatus := models.INIT_MSG
		if data.QueryCount+1 >= appInfo.Query_times {
			upStatus = models.UNCONFIRM_MSG
			content = "该事务消息未进行confirm，"
			needAlarm = true
		}
		model.ModifyQuery(data.Mid, resp.Msg, upStatus)

		// 记录日志
		logStr := "get:" + appInfo.Query_url
		if err != nil {
			logStr += ",error:" + err.Error()
		} else {
			logStr += ",response:" + resp.Result + "|" + resp.Msg
		}
		log.Error(logStr)
	case models.NOTIFYING_MSG:
		// 消息通知
		needAlarm, _ = baseObj.NotifyMessage(data)
	case models.UNCONFIRM_MSG:
		content = "该事务消息未进行confirm，"
		needAlarm = true
	case models.FAIL_MSG:
		needAlarm = true
	}

	if needAlarm {
		content += "\n mid: " + strconv.FormatInt(data.Mid, 10) + "，\n 消息内容：" + data.List
		if err := notifyHuman(content); err == nil {
			model.ModifyAlarm(data.Mid)
		}
	}
}

/**
 * 扫描消息表状态
 */
func scan() {
	defer func() {
		if err := recover(); err != nil {
			log.ErrorStrace(err)
		}
	}()
	var (
		appInfo     models.Application
		ok          bool
		isSubtable  bool = false
		runNotifyCh      = make(chan bool, server.GetConfig().Con_notices)
	)
	defaultInterval, _ := strconv.Atoi(server.GetConfig().Default_interval)
	//appTable := middleware.GetAppTable()
	model := new(models.MessageList)
	appModel := new(models.Application)
	msgC := middleware.GetMsgChannel()

	model.HandleOverflowTable()

	for {
		sleepTime := defaultInterval
		mLists := model.GetNotifyMessage(server.GetConfig().Default_interval, server.GetConfig().Alarm_interval)
		//
		for _, data := range mLists {
			if data.ID > model.GetMaxTableSize() {
				isSubtable = true
			}
			// 无需重复查表
			if appInfo, ok = middleware.GetAppTable(data.AppName); !ok {
				appInfo = appModel.GetInfosByName(data.AppName)
				middleware.SetAppTable(data.AppName, appInfo)
			}

			if data.Need_run > 0 {
				// 不需运行，取最小
				if 0 == sleepTime {
					sleepTime = data.Need_run
				} else if data.Need_run < sleepTime {
					sleepTime = data.Need_run
				}
				continue
			}

			runNotifyCh <- true
			go RunQueryNotify(data, appInfo, runNotifyCh)
		}

		if len(mLists) > 0 {
			select {
			case msgC <- true:
				time.Sleep(time.Duration(sleepTime) * time.Second)
			default:
				time.Sleep(time.Duration(sleepTime) * time.Second)
			}
		} else if isSubtable {
			model.HandleOverflowTable()
		}

		<-msgC
	}
}
