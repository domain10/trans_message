package middleware

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"trans_message/middleware/server"
	"trans_message/models"
	"trans_message/my_vendor/utils"
)

var (
	//
	lastTimeStamp int64 // 上次的时间戳(秒，占31)
	serverId      int64 // 机器 id 占10位, 十进制范围是 [ 0, 1023 ]
	sn            int64 // 序列号占 22 位,十进制范围是 [ 0, 4194303 ]
	msgChannel    chan bool
	appTable      sync.Map //map[string]models.Application
	msgHtable     sync.Map

	httpTr *http.Transport
)

type responseResult struct {
	Result string
	Msg    string
	Data   interface{}
}

type msgCacheResult struct {
	stime int64
	Data  interface{}
}

func init() {
	lastTimeStamp = time.Now().Unix()
	serverId = server.GetConfig().SERVER_ID
	// 左移22位,让出空间给序列号使用
	serverId = serverId << 22
	msgChannel = make(chan bool, 1)
	// 应用数据
	//appTable = make(map[string]models.Application)
	aLists := new(models.Application).Lists()
	for _, data := range aLists {
		//appTable[data.Name] = data
		appTable.Store(data.Name, data)
	}

	httpTr = &http.Transport{
		Dial: func(netw, addr string) (conn net.Conn, err error) {
			conn, err = net.DialTimeout(netw, addr, time.Second*5) //设置建立连接超时
			return
		},
		ResponseHeaderTimeout: time.Second * 5, //设置发送、接受头部数据超时
		MaxIdleConns:          50,
		IdleConnTimeout:       30 * time.Second,
	}
}

func GetMsgChannel() chan bool {
	return msgChannel
}

func GetAppTable(name string) (appInfo models.Application, ok bool) {
	//appInfo, ok = appTable[name]
	tmpV, _ := appTable.Load(name)
	appInfo, ok = tmpV.(models.Application)
	return
}

func SetAppTable(name string, data models.Application) {
	if data.ID >= 0 {
		//appTable[name] = data
		appTable.Store(name, data)
	}
}

func ClearTimeoutMsg() bool {
	var hasData bool = false
	msgHtable.Range(func(key, tmpV interface{}) bool {
		hasData = true
		currTime := time.Now().Unix()
		msg, ok := tmpV.(msgCacheResult)
		if ok && msg.stime+30 <= currTime {
			msgHtable.Delete(key)
		}
		return true
	})

	return hasData
}

func SetMsgHtable(key string, val interface{}, flag string) (res bool) {
	var msg msgCacheResult
	var oldval interface{}
	res = true
	msg.stime = time.Now().Unix()
	msg.Data = val

	if flag == "NX" {
		oldval, res = msgHtable.LoadOrStore(key, msg)
		if res {
			if oldMsg, ok := oldval.(msgCacheResult); ok && oldMsg.stime+30 <= time.Now().Unix() {
				msgHtable.Store(key, msg)
				res = false
			}
		}
	} else {
		msgHtable.Store(key, msg)
	}
	return
}

func DelMsgHtable(key string) {
	msgHtable.Delete(key)
}

func GenerateId() int64 {
	var id int64
	curTimeStamp := time.Now().Unix()
	// 同一秒
	if curTimeStamp <= lastTimeStamp {
		sn++
		if sn > 4194303 {
			time.Sleep(time.Second)
			curTimeStamp = time.Now().Unix()
			lastTimeStamp = curTimeStamp
			sn = 0
		}
	} else {
		sn = 0
		lastTimeStamp = curTimeStamp
	}
	// 并结果，第32位必然是0，低31位也就是时间戳的低31位
	// 机器 id 占用10位空间,序列号占用22位空间,所以左移 32 位; 经过下面的与操作,左移后的第 1 位必然是0
	rightBinValue := curTimeStamp & 0x7FFFFFFF
	rightBinValue <<= 32
	id = rightBinValue | serverId | sn
	return id
}

func GenerateAppkey() string {
	id := GenerateId()
	return utils.Md5(strconv.FormatInt(id, 10))
}

func IpAddrCheck(addr string) bool {
	if addr == "127.0.0.1" {
		return true
	}
	var result bool = false
	a := net.ParseIP(addr)
	if a == nil {
		result = false
	} else if ip4 := a.To4(); ip4 != nil {
		switch {
		case ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31:
			result = true
		default:
			result = false
		}
	}
	return result
}

/**
 * @支持get方式，post的json数据
 */
func RawRequest(url, methond, data, contentType string) (result []byte, err error) {
	var resp *http.Response
	client := &http.Client{
		Transport: httpTr,
	}

	switch methond {
	case "GET":
		if data != "" {
			url += "?" + data
		}
		resp, err = client.Get(url)
	case "POST":
		fallthrough
	default:
		//b, err := json.Marshal(rbody)
		if contentType == "" {
			contentType = "application/json;charset=utf-8"
		}
		body := strings.NewReader(data)
		resp, err = client.Post(url, contentType, body)
	}
	if err != nil {
		return
	} else if resp != nil {
		defer resp.Body.Close()
	}

	result, err = ioutil.ReadAll(resp.Body)
	return
}

func RequestOp(url, methond, data, contentType string) (result responseResult, err error) {
	respContent, err := RawRequest(url, methond, data, contentType)
	if err == nil {
		err = json.Unmarshal(respContent, &result)
		if err != nil || (result.Result == "" && result.Msg == "") {
			result.Msg = string(respContent)
			tmp := []rune(result.Msg)
			if len(tmp) > 1024 {
				result.Msg = string(tmp[:1024])
			}
		}
	}
	return
}

func EchoLog(log ...interface{}) {
	//fmt.Print("[" + time.Now().Format("2006-01-02 15:04:05") + "] ")
	fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"] ", log)
}
