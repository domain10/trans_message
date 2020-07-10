package models

import (
	"fmt"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
)

var tablesize int64 = 20000000
var tablename = "message_list"

type MessageList struct {
	ID          int64  `gorm:"column:id;primary_key"`
	Mid         int64  `gorm:"column:mid"`
	AppKey      string `gorm:"column:app_key"`
	FromAddress string `gorm:"column:from_address"`
	MessageType int    `gorm:"column:message_type"`
	List        string `gorm:"column:list"`
	Status      int    `gorm:"column:status;default:1"`
	Describe    string `gorm:"column:describe"`
	NotifyCount int    `gorm:"column:notify_count"`
	CreateTime  int    `gorm:"column:create_time"`
	UpdateTime  int    `gorm:"column:update_time"`
}

func (_ *MessageList) GetDataByid(id string) MessageList {
	var data MessageList
	if mid, err := strconv.ParseInt(id, 10, 64); err == nil {
		GetOrm(mid, nil).Select("mid,message_type,list,notify_count").Where("mid=?", mid).First(&data)
	}
	return data
}

func (self *MessageList) Insert(mid int64, appkey, ip string, message_type int, list string) int64 {
	if mid == 0 {
		return 0
	}
	strInt64 := strconv.FormatInt(time.Now().Unix(), 10)
	now, _ := strconv.Atoi(strInt64)
	result := &MessageList{Mid: mid, AppKey: appkey, FromAddress: ip, MessageType: message_type, List: list, CreateTime: now, UpdateTime: now}
	GetOrm(mid, nil).Create(result)
	if result.ID > tablesize {
		self.SubTable(mid)
	}
	return result.ID
}

func (_ *MessageList) Modify(mid int64, list, describe string, status int) (rows int64) {
	if status <= 0 {
		return 0
	}
	now := time.Now().Unix()
	tmp := map[string]interface{}{"status": status, "notify_count": gorm.Expr("notify_count+1"), "describe": describe, "update_time": now}
	if list != "" {
		tmp["list"] = list
	}
	rows = GetOrm(mid, nil).Table(tablename).Where("mid = ?", mid).Updates(tmp).RowsAffected
	fmt.Println(rows)
	return rows
}

func (_ *MessageList) GetNotifyMessage() []MessageList {
	var data []MessageList
	//ctime := time.Now().Unix() - 30
	for _, orm := range ormArr {
		var tmpData []MessageList
		orm.Select("mid,app_key,from_address,message_type,list,status,notify_count").Where("(status=1 and notify_count<=? and create_time<UNIX_TIMESTAMP()-30) or status in (2,4)", 10).Find(&tmpData)
		data = append(data, tmpData...)
	}
	return data
}

func (_ *MessageList) SubTable(mid int64) bool {
	cstZone := time.FixedZone("CST", 8*3600)
	currDate := time.Now().In(cstZone).Format("20060102")
	backupTable := tablename + "_" + currDate
	sql := "RENAME TABLE " + tablename + " TO " + backupTable + ";"
	sql2 := "CREATE TABLE IF NOT EXISTS " + tablename + " LIKE " + backupTable + ";"
	GetOrm(mid, nil).Exec(sql)
	GetOrm(mid, nil).Exec(sql2)
	return true
}
