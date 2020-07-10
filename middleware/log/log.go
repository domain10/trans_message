package log

import (
	"fmt"
	"io"
	"log"
	"strings"

	"os"
	"runtime/debug"
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
	ErrorWriter      io.Writer
	ServiceErrWriter io.Writer
	InfoWriter       io.Writer
	dateStr          string
	accessLog        *os.File
	errorLog         *os.File
	ServiceErrLog    *os.File
	infoLog          *os.File
	cstZone          = time.FixedZone("CST", 8*3600)
)

const (
	LeveL_WARNING = "warning"
	LeveL_INFO    = "info"
	LeveL_DEBUG   = "debug"
	LeveL_ERROR   = "error"
	LeveL_SERIOUS = "serious"
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
	InitAllLogger()
}

func InitAllLogger() {
	//----------M-------------
	currDate := time.Now().In(cstZone).Format("2006-01-02")
	if oldTime, err := time.ParseInLocation("2006-01-02 15:04:05", dateStr+" 00:00:00", cstZone); err == nil {
		currTime, _ := time.ParseInLocation("2006-01-02 15:04:05", currDate+" 00:00:00", cstZone)
		if currTime.Unix() > oldTime.Unix() {
			dateStr = currDate
		} else {
			return
		}
		// } else {
		// 	return
	}
	//-----------------------
	// init access.log
	if server.GetConfig().ACCESS_LOG != "" {
		gin.DefaultWriter = InitLogger(server.GetConfig().ACCESS_LOG, 1)
	}

	// init error.log
	if server.GetConfig().ERROR_LOG != "" {
		ErrorWriter = InitLogger(server.GetConfig().ERROR_LOG, 2)
	}

	if server.GetConfig().SERVICE_ERR_LOG != "" {
		ServiceErrWriter = InitLogger(server.GetConfig().SERVICE_ERR_LOG, 3)
	}
	// init info.log
	if server.GetConfig().INFO_LOG != "" {
		InfoWriter = InitLogger(server.GetConfig().INFO_LOG, 4)
	}
}

func InitLogger(path string, t int) io.Writer {
	var err error
	var filePtr *os.File
	if strings.HasSuffix(path, "/") {
		path += dateStr + ".log"
	} else {
		path += "/" + dateStr + ".log"
	}
	switch t {
	case 1:
		if accessLog != nil {
			accessLog.Close()
		}
		accessLog, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		filePtr = accessLog
	case 2:
		if errorLog != nil {
			errorLog.Close()
		}
		errorLog, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		filePtr = errorLog
	case 3:
		if ServiceErrLog != nil {
			ServiceErrLog.Close()
		}
		ServiceErrLog, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		// os.Stdout = ServiceErrLog
		// os.Stderr = ServiceErrLog
		filePtr = ServiceErrLog
	default:
		if infoLog != nil {
			infoLog.Close()
		}
		infoLog, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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
	if server.GetConfig().ERROR_LOG != "" {
		InitAllLogger()
		//-----------------------
		fmt.Fprintf(ErrorWriter, "%s", "["+time.Now().Format("2006-01-02 15:04:05")+"] ")
		fmt.Fprintf(ErrorWriter, "%s", data)
		fmt.Fprintf(ErrorWriter, "%s", "\n")
	}
}
func ErrorStrace(data interface{}) {
	if server.GetConfig().SERVICE_ERR_LOG != "" {
		InitAllLogger()
		//-----------------------
		fmt.Fprintf(ServiceErrWriter, "%s", "["+time.Now().Format("2006-01-02 15:04:05")+"] ")
		fmt.Fprintf(ServiceErrWriter, "%s", data)
		fmt.Fprintf(ServiceErrWriter, "%s", "\n")
		fmt.Fprintf(ServiceErrWriter, "%s", "Stack trace:\n")
		fmt.Fprintf(ServiceErrWriter, "%s", debug.Stack())
		fmt.Fprintf(ServiceErrWriter, "%s", "\n")
	}
}

func Info(info E) {
	if server.GetConfig().INFO_LOG != "" {
		InitAllLogger()
		//-----------------------
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
