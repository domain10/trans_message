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
	"trans_message/my_vendor/utils"
	//"github.com/gin-gonic/gin"
)

type Base struct {
}

type MessageData struct {
	To_app  string `json:"to_app"`
	To_url  string `json:"to_url"`
	Content string `json:"content"`
	Status  int    `json:"status"`
}

type NotifyData struct {
	Mid     int64  `json:"mid"`
	Content string `json:"content"`
	Time    int64  `json:"time"`
	Sign    string `json:"sign"`
}

/**
 * 发送通知事务消息
 */
func (_ *Base) NotifyMessage(data models.MessageList) (needAlarm bool, err error) {
	var list []MessageData

	if err = json.Unmarshal([]byte(data.List), &list); err == nil {
		var (
			i          int = 0
			postData   NotifyData
			detail     string = ""
			maxNotices int    = 10
			appkey     string
		)
		model := new(models.MessageList)
		for k, v := range list {
			// 获取通知该应用的信息
			if toApp, ok := middleware.GetAppTable(v.To_app); ok {
				maxNotices = toApp.Notify_times
				appkey = toApp.App_key
			}

			if v.Status == 0 {
				postData.Mid = data.Mid
				postData.Content = v.Content
				postData.Time = time.Now().Unix()
				tmpStr := strconv.FormatInt(data.Mid, 10) + "|" + strconv.FormatInt(postData.Time, 10) + "|" + appkey
				postData.Sign = utils.HmacSha256(tmpStr, appkey, true)
				postStr, _ := json.Marshal(postData)

				resp, err := middleware.RequestOp(v.To_url, "POST", string(postStr), "")
				detail += resp.Result + "," + resp.Msg + ";"

				if err == nil && "success" == resp.Result {
					i++
					list[k].Status = 1
				} else {
					logStr := "post:" + v.To_url + ",data:" + string(postStr)
					if err != nil {
						logStr += ",error:" + err.Error() + "|" + resp.Msg
					} else {
						logStr += ",response:" + resp.Result + "|" + resp.Msg
					}
					log.Error(logStr)
					//依次执行的消息，本消息没成功，则后面的不能发送
					if 2 == data.MessageType {
						break
					}
				}
			} else {
				i++
			}
		}

		tmpContent, _ := json.Marshal(list)
		if len(list) == i {
			// 成功
			model.ModifyNotify(data.Mid, string(tmpContent), detail, models.SUCCESS_MSG)
		} else if data.NotifyCount+1 >= maxNotices {
			// 达到最大通知次数
			model.ModifyNotify(data.Mid, string(tmpContent), detail, models.FAIL_MSG)
			needAlarm = true
		} else {
			// 新消息、已通知过还未成功的
			model.ModifyNotify(data.Mid, string(tmpContent), detail, models.NOTIFYING_MSG)
		}
	}

	err = nil
	return
}

func (_ *Base) GetAppInfos(name string) (res models.Application) {
	appModel := new(models.Application)
	res = appModel.GetInfosByName(name)
	return
}

func (self *Base) CheckAppExist(name string) (res bool) {
	res = false
	appInfo := self.GetAppInfos(name)
	if appInfo.ID > 0 {
		res = true
	}
	return
}

func (_ *Base) CheckDuplicated(data string) (res bool) {
	res = true
	cacheObj := new(cache.Handler).On()
	tmp, err := cacheObj.Set("set:check_message:"+utils.Md5(data), data, "EX", 30, "NX")
	//fmt.Println(tmp, ",", err)
	if err == nil && strings.ToLower(tmp) == "ok" {
		res = false
	}
	return
}

/*
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
*/
