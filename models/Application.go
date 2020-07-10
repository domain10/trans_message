package models

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"
	"trans_message/models/cache"
)

var hkey = "hash:application:tab"

type Application struct {
	ID         int    `gorm:"column:id;primary_key"`
	Name       string `gorm:"column:name"`
	App_key    string `gorm:"app_key"`
	Ip         string `gorm:"column:ip"`
	Query_url  string `gorm:"column:query_url"`
	Notify_url string `gorm:"column:notify_url"`
	Describe   string `gorm:"column:describe"`
	CreateTime int    `gorm:"column:create_time"`
	UpdateTime int    `gorm:"column:update_time"`
}

// TableName
func (_ *Application) TableName() string {
	return "application"
}

func (data *Application) GetInfosByAppkey(appkey string) *Application {
	cacheObj := new(cache.Handler).On()
	if tmp, err := cacheObj.Hget(hkey, appkey); err == nil && tmp != "" {
		err = json.Unmarshal([]byte(tmp), data)
	} else {
		GetOrm(0, nil).Select("id,name,app_key,ip,query_url,notify_url").Where("app_key=? and status=1", appkey).First(data)
		if tmp, err := json.Marshal(data); err == nil {
			cacheObj.Hset(hkey, appkey, string(tmp))
		}
	}
	return data
}

// func (_ *Application) Lists() []Application {
// 	var data []Application
// 	GetOrm(0, nil).Select("id,name,app_key,ip,query_url,notify_url").Where("status=1").Find(&data)
// 	return data
// }

func (self *Application) RegisterOp(data *Application) (err error) {
	if data.App_key == "" {
		err = errors.New("Registration failure: no key was generated")
		return
	}
	var existed Application
	tx := GetOrm(0, nil).Begin()
	if err = tx.Error; err != nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	tx.Raw("select id from "+self.TableName()+" where name = ? for update", data.Name).Scan(&existed)
	if existed.ID == 0 {
		strInt64 := strconv.FormatInt(time.Now().Unix(), 10)
		now, _ := strconv.Atoi(strInt64)
		data.CreateTime = now
		data.UpdateTime = now
		err = tx.Create(data).Error
		if err == nil {
			err = tx.Commit().Error
			//
			tmp, tmpErr := json.Marshal(data)
			if err == nil && tmpErr == nil {
				new(cache.Handler).On().Hset(hkey, data.App_key, string(tmp))
			}
		} else {
			tx.Rollback()
		}
	} else {
		tx.Rollback()
		err = errors.New("The application name already exists")
	}
	return
}
