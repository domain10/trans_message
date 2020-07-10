package main

import (
	"strconv"
	"time"
	"trans_message/controllers"
	"trans_message/middleware"
	"trans_message/middleware/log"
	"trans_message/middleware/server"
	"trans_message/models"
)

/**
 * @扫描消息表状态
 */
func scan() {
	defer func() {
		if err := recover(); err != nil {
			log.ErrorStrace(err)
		}
	}()
	appTable := make(map[string]*models.Application)
	model := new(models.MessageList)
	appModel := new(models.Application)
	baseObj := new(controllers.Base)
	model.SubTable(0)
	for {
		mLists := model.GetNotifyMessage()
		for _, data := range mLists {
			var appInfo *models.Application
			var ok bool
			midStr := strconv.FormatInt(data.Mid, 10)
			appInfo, ok = appTable[data.AppKey]
			if !ok {
				appInfo = appModel.GetInfosByAppkey(data.AppKey)
				appTable[data.AppKey] = appInfo
			}
			//
			switch data.Status {
			case 1:
				//查询确认消息
				respResult, err := middleware.RequestOp(appInfo.Query_url, "GET", "mid="+midStr, "")
				if err == nil && "submit" == respResult {
					baseObj.NotifyMessage(data)
				} else {
					model.Modify(data.Mid, "", respResult, 1)
					logStr := "get:" + appInfo.Query_url + ",response:" + respResult
					if err != nil {
						logStr += ",error:" + err.Error()
					}
					log.Error(logStr)
				}
			case 2:
				//消息通知
				baseObj.NotifyMessage(data)
			case 4:
				//通知人工处理
				content := "消息mid：" + midStr + "，消息内容：" + data.List
				postData := "title=事务消息通知失败&content=" + content + "&ding_userid=" + server.GetConfig().Ding_userid
				respResult, err := middleware.RequestOp(server.GetConfig().Ding_news_url, "POST", postData, "application/x-www-form-urlencoded;charset=utf-8")
				if err == nil {
					model.Modify(data.Mid, "", respResult, 5)
				} else {
					log.Error("通知人工处理出错,mid:" + midStr + ",post:" + postData + ",response:" + respResult + ",error:" + err.Error())
				}
			}
		}
		if len(mLists) > 0 {
			time.Sleep(60 * time.Second)
		} else {
			time.Sleep(10 * time.Second)
		}
	}
}
