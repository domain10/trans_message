package core

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Context struct {
	*gin.Context
}

func (c *Context) Success(data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"msg":  "success",
		"data": data,
	})
}

func (c *Context) Fail(retcode int, msg string, data interface{}) {
	c.JSON(retcode, gin.H{
		"msg":  msg,
		"data": data,
	})
}

// func (c *Context) ErrorPage(code int, msg, data string) {
// 	c.HTML(http.StatusOK, "error/error", gin.H{
// 		"code": code,
// 		"msg":  msg,
// 		"data": data
// 	})
// }
