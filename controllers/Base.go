package controllers

import (
	"encoding/json"

	"strconv"
	"strings"
	"time"
	"trans_message/middleware"
	"trans_message/middleware/log"
	"trans_message/models"
	"trans_message/models/cache"
	"utils"
	//"github.com/gin-gonic/gin"
)

type Base struct {
}

type MessageData struct {
	To_url  string `json:"to_url"`
	Content string `json:"content"`
	Status  int    `json:"status"`
}

func (_ *Base) NotifyMessage(data models.MessageList) (err error) {
	//data.Mid, data.MessageType, data.List, data.NotifyCount
	var list []MessageData
	var respResult string
	model := new(models.MessageList)
	mid := strconv.FormatInt(data.Mid, 10)
	if err = json.Unmarshal([]byte(data.List), &list); err == nil {
		var i int = 0
		for k, v := range list {
			if v.Status == 0 {
				postData := "{\"mid\":" + mid + ",\"content\":\"" + v.Content + "\"}"
				respResult, err = middleware.RequestOp(v.To_url, "POST", postData, "")
				if err == nil && "success" == respResult {
					i++
					list[k].Status = 1
				} else {
					logStr := "post:" + v.To_url + ",data:" + postData + ",response:" + respResult
					if err != nil {
						logStr += ",error:" + err.Error()
					}
					log.Error(logStr)
					if 2 == data.MessageType {
						//依次执行的消息，本消息没成功，则后面的不能发送
						break
					}
				}
			} else {
				i++
			}
		}
		tmpContent, _ := json.Marshal(list)
		if len(list) == i {
			//成功
			model.Modify(data.Mid, string(tmpContent), respResult, 3)
		} else if data.NotifyCount >= 9 {
			//失败了10次，得发送钉钉通知
			model.Modify(data.Mid, string(tmpContent), respResult, 4)
		} else {
			//新消息、已通知过的
			model.Modify(data.Mid, string(tmpContent), respResult, 2)
		}
	}
	err = nil
	return
}

func (_ *Base) NotifyMessageR(data RequestData) (err error) {
	//data.Mid, data.List, data.Message_type, data.Notify_count
	mid := data.Mid
	var i int = 0
	for k, v := range data.List {
		if v.Status == 0 {
			postData := "{\"mid\":" + mid + ",\"content\":\"" + v.Content + "\"}"
			respResult, err := middleware.RequestOp(v.To_url, "POST", postData, "")
			if err == nil && "success" == respResult {
				i++
				data.List[k].Status = 1
			} else if 2 == data.Message_type {
				//依次执行的消息，本消息没成功，则后面的不能发送
				logStr := "post:" + v.To_url + ",data:" + postData + ",response:" + respResult
				if err != nil {
					logStr += ",error:" + err.Error()
				}
				log.Error(logStr)
				err = nil
				break
			} else {
				logStr := "post:" + v.To_url + ",data:" + postData + ",response:" + respResult
				if err != nil {
					logStr += ",error:" + err.Error()
				}
				log.Error(logStr)
				err = nil
			}
		} else {
			i++
		}
	}
	if len(data.List) == i {
		//成功
		data.Status = 3
	} else if data.Notify_count >= 9 {
		//通知了10次还失败
		data.Status = 4
		//转人工处理
	} else {
		//新消息、已通知过的
		data.Status = 2
	}
	data.Notify_count++
	data.Utime = time.Now().Unix()
	//tmpContent, _ := json.Marshal(data)
	// cacheObj := new(cache.Handler).On()
	// _, err = cacheObj.Modify(mid, string(tmpContent))
	return
}

func (_ *Base) CheckAppExist(appKey string) (res bool) {
	res = false
	appModel := new(models.Application)
	appInfo := appModel.GetInfosByAppkey(appKey)
	if appInfo.ID > 0 {
		res = true
	}
	return
}

func (_ *Base) CheckDuplicated(data string) (res bool) {
	res = true
	cacheObj := new(cache.Handler).On()
	tmp, err := cacheObj.Set("set:check_message", utils.Md5(data), "EX", 30, "NX")
	//fmt.Println(tmp, ",", err)
	if err == nil && strings.ToLower(tmp) == "ok" {
		res = false
	}
	return
}
