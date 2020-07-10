package middleware

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"trans_message/middleware/server"
	"utils"
)

var (
	//
	lastTimeStamp int64 // 上次的时间戳(秒，占31)
	serverId      int64 // 机器 id 占10位, 十进制范围是 [ 0, 1023 ]
	sn            int64 // 序列号占 22 位,十进制范围是 [ 0, 4194303 ]
)

type ResponseResult struct {
	Msg  string
	Data interface{}
}

func init() {
	lastTimeStamp = time.Now().Unix()
	serverId = server.GetConfig().SERVER_ID
	// 左移12位,让出空间给序列号使用
	serverId = serverId << 22
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
	// 机器 id 占用10位空间,序列号占用23位空间,所以左移 33 位; 经过下面的与操作,左移后的第 1 位必然是0
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
func RequestOp(url, methond, data, contentType string) (respResult string, err error) {
	var tmpResponse ResponseResult
	var resp *http.Response
	switch methond {
	case "GET":
		if data != "" {
			url += "?" + data
		}
		resp, err = http.Get(url)
	case "POST":
		fallthrough
	default:
		//b, err := json.Marshal(rbody)
		if contentType == "" {
			contentType = "application/json;charset=utf-8"
		}
		body := strings.NewReader(data)
		resp, err = http.Post(url, contentType, body)
	}
	if err != nil {
		return "", err
	} else if resp != nil {
		defer resp.Body.Close()
	}
	respContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(respContent, &tmpResponse)
	if err != nil || tmpResponse.Msg == "" {
		tmp := []rune(string(respContent))
		if len(tmp) > 255 {
			respResult = string(tmp[0:255])
		} else {
			respResult = string(respContent)
		}

	} else {
		respResult = tmpResponse.Msg
	}
	return
}

func EchoLog(log ...interface{}) {
	//fmt.Print("[" + time.Now().Format("2006-01-02 15:04:05") + "] ")
	fmt.Println("["+time.Now().Format("2006-01-02 15:04:05")+"] ", log)
}
