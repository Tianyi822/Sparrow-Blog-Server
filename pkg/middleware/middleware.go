package middleware

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"h2blog_server/pkg/logger"
	"net/http"
)

type RequestInfo struct {
	IP     string      `json:"ip"`
	Method string      `json:"method"`
	URL    string      `json:"url"`
	Header http.Header `json:"header"`
}

func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		// 构造我们需要的结构体
		reqInfo := RequestInfo{
			IP:     ctx.Request.Host,
			Method: ctx.Request.Method,
			URL:    ctx.Request.URL.String(),
			Header: ctx.Request.Header,
		}

		// 转为 JSON
		jsonData, err := json.Marshal(reqInfo)
		if err != nil {
			logger.Error("{\"ROUTER\": %v, \"encode_json_err\": %v}", reqInfo, err)
		} else {
			logger.Info("{\"ROUTER\": %v}", string(jsonData))
		}
		// 处理请求
		ctx.Next()
	}
}
