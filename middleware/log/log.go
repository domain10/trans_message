package log

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"runtime/debug"
	"sort"
	"strings"
	"time"
	"trans_message/middleware/server"

	"github.com/gin-gonic/gin"
)

// ---------------------------------------------------------------
// Logger
// ---------------------------------------------------------------
//
// there are three kinds of logger：
//
// 1） access logger
//     log every request visited which used to counts the ip
//     and other indicators
//
// 2） error logger
//	   record the panic error
//
// 3） info logger
//     log something the developer wants to output
//
// ---------------------------------------------------------------

var (
	ErrorWriter     io.Writer
	SystemErrWriter io.Writer
	InfoWriter      io.Writer
	dateStr         string
	infoLog         *os.File
	accessLog       *os.File
	errorLog        *os.File
	systemErrLog    *os.File
	cstZone         = time.FixedZone("CST", 8*3600)
)

const (
	LEVEL_INFO int = iota
	LEVEL_ACCESS
	LEVEL_ERROR
	LEVEL_SYSTEM_ERROR
)

type E struct {
	Function string
	Error    error
	Title    string
	Info     M
	Level    string
	Context  *gin.Context
}

type M map[string]interface{}

func init() {
	dateStr = time.Now().In(cstZone).Format("2006-01-02")
	startLog := []int{LEVEL_ERROR, LEVEL_SYSTEM_ERROR}
	for _, t := range startLog {
		InitAllLogger(t, true)
	}

}

func InitAllLogger(logType int, reopen bool) {
	// 按天生成日志文件
	currDate := time.Now().In(cstZone).Format("2006-01-02")
	if oldTime, err := time.ParseInLocation("2006-01-02 15:04:05", dateStr+" 00:00:00", cstZone); err == nil {
		currTime, _ := time.ParseInLocation("2006-01-02 15:04:05", currDate+" 00:00:00", cstZone)
		if currTime.Unix() > oldTime.Unix() {
			dateStr = currDate
		} else if !reopen {
			return
		}
	}
	switch logType {
	case LEVEL_INFO:
		// init info.log
		if server.GetConfig().INFO_LOG != "" {
			InfoWriter = InitLogger(server.GetConfig().INFO_LOG, LEVEL_INFO)
		}
	case LEVEL_ACCESS:
		// init access.log
		if server.GetConfig().ACCESS_LOG != "" {
			gin.DefaultWriter = InitLogger(server.GetConfig().ACCESS_LOG, LEVEL_ACCESS)
		}
	case LEVEL_ERROR:
		// init error.log
		if server.GetConfig().ERROR_LOG != "" {
			ErrorWriter = InitLogger(server.GetConfig().ERROR_LOG, LEVEL_ERROR)
		}
	case LEVEL_SYSTEM_ERROR:
		if server.GetConfig().Trans_system_err != "" {
			SystemErrWriter = InitLogger(server.GetConfig().Trans_system_err, LEVEL_SYSTEM_ERROR)
		}
	}
}

/**
 * 降序
 */
func sortByTime(pl []os.FileInfo) []os.FileInfo {
	sort.Slice(pl, func(i, j int) bool {
		flag := false
		if pl[i].ModTime().After(pl[j].ModTime()) {
			flag = true
		} else if pl[i].ModTime().Equal(pl[j].ModTime()) {
			if pl[i].Name() < pl[j].Name() {
				flag = true
			}
		}
		return flag
	})
	return pl
}

func clearLogFile(folder string) {
	files, errDir := ioutil.ReadDir(folder)
	if errDir != nil {
		//Println("[提示]", errDir)
		return
	}

	files = sortByTime(files)
	i := 0
	for _, file := range files {
		if !file.IsDir() {
			i++
			if i > server.GetConfig().Max_files {
				os.Remove(folder + file.Name())
			}
		}
	}
}

func InitLogger(path string, t int) io.Writer {
	var err error
	var filePtr *os.File
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	path = server.RunPath() + strings.TrimLeft(path, ".")
	// 清楚历史日志文件
	clearLogFile(path)

	path += dateStr + ".log"

	switch t {
	case LEVEL_ACCESS:
		if accessLog != nil {
			accessLog.Close()
		}
		accessLog, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
		filePtr = accessLog
	case LEVEL_ERROR:
		if errorLog != nil {
			errorLog.Close()
		}
		errorLog, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
		filePtr = errorLog
	case LEVEL_SYSTEM_ERROR:
		if systemErrLog != nil {
			systemErrLog.Close()
		}
		systemErrLog, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
		// os.Stdout = systemErrLog
		// os.Stderr = systemErrLog
		filePtr = systemErrLog
	default:
		if infoLog != nil {
			infoLog.Close()
		}
		infoLog, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
		filePtr = infoLog
	}
	if err != nil {
		log.Fatalln(err)
	}
	if server.GetConfig().DEBUG {
		return io.MultiWriter(filePtr, os.Stdout)
	} else {
		return io.MultiWriter(filePtr)
	}
}

func Error(data interface{}) {
	reopen := false
	if ErrorWriter == nil {
		reopen = true
	}
	InitAllLogger(LEVEL_ERROR, reopen)

	if ErrorWriter != nil {
		fmt.Fprintf(ErrorWriter, "%s", "["+time.Now().Format("2006-01-02 15:04:05")+"] ")
		fmt.Fprintf(ErrorWriter, "%s", data)
		fmt.Fprintf(ErrorWriter, "%s", "\n")
	}
}
func ErrorStrace(data interface{}) {
	reopen := false
	if SystemErrWriter == nil {
		reopen = true
	}
	InitAllLogger(LEVEL_SYSTEM_ERROR, reopen)

	if SystemErrWriter != nil {
		fmt.Fprintf(SystemErrWriter, "%s", "["+time.Now().Format("2006-01-02 15:04:05")+"] ")
		fmt.Fprintf(SystemErrWriter, "%s", data)
		fmt.Fprintf(SystemErrWriter, "%s", "\n")
		fmt.Fprintf(SystemErrWriter, "%s", "Stack trace:\n")
		fmt.Fprintf(SystemErrWriter, "%s", debug.Stack())
		fmt.Fprintf(SystemErrWriter, "%s", "\n")
	}
}

func Info(info E) {
	reopen := false
	if InfoWriter == nil {
		reopen = true
	}
	InitAllLogger(LEVEL_INFO, reopen)

	if InfoWriter != nil {
		fmt.Fprintf(InfoWriter, "%s", "["+time.Now().Format("2006-01-02 15:04:05")+"]")
		if info.Level == "" {
			info.Level = "info"
		}
		fmt.Fprintf(InfoWriter, "level=%s ", info.Level)

		if info.Context != nil {
			fmt.Fprintf(InfoWriter, "method=%s path=%s ", info.Context.Request.Method, info.Context.Request.URL.Path)
		}

		if info.Function != "" {
			fmt.Fprintf(InfoWriter, "function=%s ", info.Function)
		}

		if info.Title != "" {
			fmt.Fprintf(InfoWriter, "title=%s ", info.Title)
		}

		if info.Error != nil {
			fmt.Fprintf(InfoWriter, "error=%s ", info.Error)
		}

		for k, v := range info.Info {
			fmt.Fprintf(InfoWriter, "%s=%v ", k, v)
		}
		fmt.Fprintf(InfoWriter, "\n")
	}
}

func Println(a ...interface{}) {
	if server.GetConfig().DEBUG {
		fmt.Println(a...)
	}
}

func Printf(format string, a ...interface{}) {
	if server.GetConfig().DEBUG {
		fmt.Printf(format, a...)
	}
}
