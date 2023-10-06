package models

import (
	"errors"
	// "encoding/json"
	//"math/rand"
	// "strconv"
	// "time"
	// "trans_message/models/cache"
)

// var hkey = "table:application:name"

type Application struct {
	ID              int    `gorm:"column:id;primary_key"`
	Name            string `gorm:"column:name"`
	App_key         string `gorm:"app_key"`
	Ip_arr          string `gorm:"column:ip_arr"`
	Query_url       string `gorm:"column:query_url"`
	Query_times     int    `gorm:"column:query_times;default:10"`
	Query_interval  int    `gorm:"column:query_interval;default:10"`
	Notify_url      string `gorm:"column:notify_url"`
	Notify_times    int    `gorm:"column:notify_times;default:0"`
	Notify_interval int    `gorm:"column:notify_interval;default:10"`
	Describe        string `gorm:"column:describe"`
}

// TableName
func (_ *Application) TableName() string {
	return "application"
}

func (_ *Application) GetInfosByName(name string) Application {
	var data Application
	if name == "" {
		return data
	}
	GetOrm(0, nil).Where("name=? and status=1", name).First(&data)
	return data
	//cacheObj := new(cache.Handler).On()
	//if tmp, err := cacheObj.Hget(hkey, name); err == nil && tmp != "" {
	//	err = json.Unmarshal([]byte(tmp), &data)
	//} else {
	// GetOrm(0, nil).Where("name=? and status=1", name).First(&data)
	// if tmp, err := json.Marshal(data); err == nil {
	// 	cacheObj.Hset(hkey, name, string(tmp))
	// 	cacheObj.Expire(hkey, 12*3600+rand.Intn(300))
	// }
	//}

}

func (_ *Application) Lists() []Application {
	var data []Application
	GetOrm(0, nil).Select("id,name,app_key,query_url,query_times,query_interval,notify_url,notify_times,notify_interval").Where("status=1").Find(&data)
	return data
}

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
		// strInt64 := strconv.FormatInt(time.Now().Unix(), 10)
		// now, _ := strconv.Atoi(strInt64)
		// data.CreateTime = now
		// data.UpdateTime = now
		err = tx.Create(data).Error
		if err == nil {
			err = tx.Commit().Error
			//
			// tmp, tmpErr := json.Marshal(data)
			// if err == nil && tmpErr == nil {
			// 	new(cache.Handler).On().Hset(hkey, data.App_key, string(tmp))
			// }
		} else {
			tx.Rollback()
		}
	} else {
		tx.Rollback()
		err = errors.New("The application name already exists")
	}
	return
}
