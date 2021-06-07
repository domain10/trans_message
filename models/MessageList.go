package models

import (
	//"encoding/json"
	// "fmt"
	"time"

	//"trans_message/models/cache"

	"github.com/jinzhu/gorm"
)

var (
	scanLimit int   = 100
	maxsize   int64 = 50000000
	// 详情字段长度
	detailLen   int = 255
	tablename       = "message_list"
	subtablekey     = "subtable:message_lisg:tag"
)

const (
	INIT_MSG int = iota
	UNCONFIRM_MSG
	NOTIFYING_MSG
	CANCEL_MSG
	FAIL_MSG
	SUCCESS_MSG
	HUMAN_MSG
)

type cacheData struct {
	Mid        int64  `json:"mid"`
	CreateTime string `json:"create_time"`
}

type MessageList struct {
	ID          int64  `gorm:"column:id;primary_key"`
	Mid         int64  `gorm:"column:mid"`
	AppName     string `gorm:"column:app_name"`
	FromAddress string `gorm:"column:from_address"`
	MessageType int    `gorm:"column:message_type"`
	List        string `gorm:"column:list"`
	Status      int    `gorm:"column:status;default:0"`
	Describe    string `gorm:"column:describe"`
	QueryCount  int    `gorm:"column:query_count"`
	NotifyCount int    `gorm:"column:notify_count"`
	AlarmCount  int    `gorm:"column:alarm_count"`
	Need_run    int    `sql:"-"`
}

func (_ *MessageList) GetMaxTableSize() int64 {
	return maxsize
}

func (_ *MessageList) GetDataBymid(mid int64) MessageList {
	var data MessageList
	GetOrm(mid, nil).Select("id,mid,message_type,list,status,notify_count").Where("mid=?", mid).First(&data)
	return data
}

func (self *MessageList) Insert(mid int64, name, ip string, message_type int, list string) int64 {
	if mid == 0 {
		return 0
	}
	result := &MessageList{Mid: mid, AppName: name, FromAddress: ip, MessageType: message_type, List: list}
	GetOrm(mid, nil).Create(result)
	return result.ID

	// strInt64 := strconv.FormatInt(time.Now().Unix(), 10)
	// now, _ := strconv.Atoi(strInt64)
	// var data cacheData
	// cstZone := time.FixedZone("CST", 8*3600)
	// data.Mid = mid
	// data.CreateTime = time.Now().In(cstZone).Format("2006-01-02 15:04:05")

	// cacheObj := new(cache.Handler).On()
	// jdata, _ := json.Marshal(data)
	// cacheObj.Set(subtablekey, string(jdata))

}

/**
 * 修改查询确认次数
 */
func (_ *MessageList) ModifyQuery(mid int64, describe string, status int) (rows int64) {
	tmpVal := []rune(describe)
	if len(tmpVal) > detailLen {
		describe = string(tmpVal[:detailLen])
	}
	tmp := map[string]interface{}{"status": status, "query_count": gorm.Expr("query_count+1"), "describe": describe}
	rows = GetOrm(mid, nil).Table(tablename).Where("mid = ? and status <= ?", mid, status).Updates(tmp).RowsAffected
	return rows
}

/**
 * 修改通知次数
 */
func (_ *MessageList) ModifyNotify(mid int64, list, describe string, status int) (rows int64) {
	tmpVal := []rune(describe)
	if len(tmpVal) > detailLen {
		describe = string(tmpVal[:detailLen])
	}
	tmp := map[string]interface{}{"status": status, "notify_count": gorm.Expr("notify_count+1"), "describe": describe}
	if list != "" {
		tmp["list"] = list
	}
	rows = GetOrm(mid, nil).Table(tablename).Where("mid = ? and status <= ?", mid, status).Updates(tmp).RowsAffected
	return rows
}

/**
 * 修改告警次数
 */
func (_ *MessageList) ModifyAlarm(mid int64) (rows int64) {
	tmp := map[string]interface{}{"alarm_count": gorm.Expr("alarm_count+1")}
	rows = GetOrm(mid, nil).Table(tablename).Where("mid = ?", mid).Updates(tmp).RowsAffected
	return rows
}

/**
 * 修改状态
 */
func (_ *MessageList) ModifyStatus(mid int64, list, describe string, status int) (rows int64) {
	tmp := map[string]interface{}{"status": status, "describe": describe}
	if list != "" {
		tmp["list"] = list
	}
	if status == FAIL_MSG {
		tmp["alarm_count"] = 0
	}

	rows = GetOrm(mid, nil).Table(tablename).Where("mid = ? and status <= ?", mid, status).Updates(tmp).RowsAffected
	return rows
}

/**
 * 更新描述
 */
func (_ *MessageList) UpdateDescribe(mid int64, describe string) (rows int64) {
	tmp := map[string]interface{}{"describe": describe}
	rows = GetOrm(mid, nil).Table(tablename).Where("mid = ?", mid).Updates(tmp).RowsAffected
	return rows
}

/**
 * 查询待处理的消息
 */
func (_ *MessageList) GetNotifyMessage(interval, alarm_interval string) []MessageList {
	var data []MessageList
	sTime := 7 * 24 * 3600
	fields := "id,mid,app_name,message_type,list,status,query_count,notify_count,alarm_count"
	fields += ",if(`status`=1 or `status`=4,unix_timestamp(update_time)+" + alarm_interval + "-unix_timestamp(),unix_timestamp(update_time)+" + interval + "-unix_timestamp()) as need_run"
	where := "create_time >=FROM_UNIXTIME(UNIX_TIMESTAMP()-?) and status in (0,1,2,4)"

	for _, orm := range ormArr {
		var tmpData []MessageList
		orm.Table(tablename).Select(fields).Where(where, sTime).Limit(scanLimit).Find(&tmpData)
		data = append(data, tmpData...)
	}
	return data
	// where := "(status=0 and create_time<from_unixtime(unix_timestamp()-?)) or (status=2 and update_time<from_unixtime(unix_timestamp()-?)) "
	// where += "or (status in (1,4) and update_time<from_unixtime(unix_timestamp()-?))"

	//
	// cacheObj := new(cache.Handler).On()
	// var subData cacheData
	// if tmp, err := cacheObj.Get(subtablekey); err == nil && tmp != "" {
	// 	var tmpData []MessageList
	// 	err = json.Unmarshal([]byte(tmp), &subData)
	// 	GetOrm(subData.Mid, nil).Table(subData.BackupName).Select(fields).Where(where, interval, interval, alarm_interval).Limit(scanLimit).Find(&tmpData)
	// 	if len(tmpData) > 0 {
	// 		data = append(data, tmpData...)
	// 	} else {
	// 		cacheObj.Del(subtablekey)
	// 	}
	// }
}

func (self *MessageList) HandleOverflowTable() (midArr []int64) {
	var count int64
	var data int64
	for _, orm := range ormArr {
		orm.Raw("select count(*) as count from " + tablename).Row().Scan(&count)
		if count >= maxsize {
			orm.Raw("select mid from " + tablename + " order by id desc limit 1").Row().Scan(&data)
			self.SubTable(data)
		}
	}
	return
}

/**
 * 自动切表
 */
func (_ *MessageList) SubTable(mid int64) bool {
	if mid < 0 {
		return false
	}
	cstZone := time.FixedZone("CST", 8*3600)
	currDate := time.Now().In(cstZone).Format("20060102")
	backupName := tablename + "_" + currDate

	sql := "RENAME TABLE " + tablename + " TO " + backupName + ";"
	sql2 := "CREATE TABLE IF NOT EXISTS " + tablename + " LIKE " + backupName + ";"
	db := GetOrm(mid, nil)
	db.Exec(sql)
	db.Exec(sql2)

	return true
}
