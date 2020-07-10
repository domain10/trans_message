package middleware

import (
	"core"
	"net/http"
	"trans_message/middleware/log"

	"github.com/go-sql-driver/mysql"
)

func Auth(c *core.Context) {
	ip := c.ClientIP()
	if !IpAddrCheck(ip) {
		c.Fail(http.StatusUnauthorized, "No access rights", nil)
		c.Abort()
	}
	c.Next()
}

func Nonexistent(c *core.Context) {
	c.Fail(http.StatusNotFound, "There is no such route or method", nil)
	return
}

func HandleErrors(c *core.Context) {
	defer func() {
		if err := recover(); err != nil {
			log.ErrorStrace(err)
			var (
				errMsg     string
				mysqlError *mysql.MySQLError
				ok         bool
			)
			if errMsg, ok = err.(string); ok {
				c.Fail(http.StatusInternalServerError, "system error, "+errMsg, nil)
				return
			} else if mysqlError, ok = err.(*mysql.MySQLError); ok {
				c.Fail(http.StatusInternalServerError, "system error, "+mysqlError.Error(), nil)
				return
			} else {
				c.Fail(http.StatusInternalServerError, "system error", nil)
				return
			}
		}
	}()
	c.Next()
}
