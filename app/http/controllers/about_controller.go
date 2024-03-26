package controllers

import (
	"github.com/leancodebox/GooseForum/app/http/controllers/component"
	"github.com/leancodebox/GooseForum/bundles/logging"
	"github.com/leancodebox/GooseForum/bundles/serverinfo"
	"net/http"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/spf13/cast"
)

func Api(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]any{
		"msg": "OK",
	})
}

const (
	contentTypeHTML      = "text/html"
	errorCodeNotFound    = 404
	errorMessageNotFound = "路由未定义，请确认 url 和请求方法是否正确。"
)

func NotFound(c *gin.Context) {
	acceptString := c.GetHeader("Accept")
	if strings.Contains(acceptString, contentTypeHTML) {
		c.Redirect(http.StatusTemporaryRedirect, "/actor")
		return
	}
	c.JSON(http.StatusNotFound, component.DataMap{
		"error_code":    errorCodeNotFound,
		"error_message": errorMessageNotFound,
	})
}

func GetUseMem() component.Response {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return component.SuccessResponse(cast.ToString(m.Alloc/1024) + "KB")
}

func About() component.Response {
	return component.SuccessResponse(component.DataMap{
		"message": "Hello~ Now you see a json from gin",
	})
}

func SysInfo() component.Response {
	var s serverinfo.Server
	var err error
	s.Os = serverinfo.InitOS()
	if s.Cpu, err = serverinfo.InitCPU(); err != nil {
		logging.ErrIf(err)

	}

	if s.Ram, err = serverinfo.InitRAM(); err != nil {
		logging.ErrIf(err)
	}

	if s.Disk, err = serverinfo.InitDisk(); err != nil {
		logging.ErrIf(err)
	}

	return component.SuccessResponse(s)
}
