package middleware

import (
	"bytes"
	"dialogTree/service/access_service"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

// AccessResponseWriter 自定义ResponseWriter来捕获响应内容
type AccessResponseWriter struct {
	gin.ResponseWriter
	Body []byte
	Head http.Header
}

func (w *AccessResponseWriter) Write(b []byte) (int, error) {
	// 拦截响应内容并存储
	w.Body = append(w.Body, b...)
	// 继续调用原来的方法
	return w.ResponseWriter.Write(b)
}

func (w *AccessResponseWriter) Header() http.Header {
	return w.Head
}

func AccessLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 处理请求体的阅后即焚特性
		byteData, err := c.GetRawData()
		if err == nil {
			// 重新写回请求体，确保后续处理可以正常使用
			c.Request.Body = io.NopCloser(bytes.NewReader(byteData))
		}

		// 创建自定义ResponseWriter来捕获响应
		resWriter := &AccessResponseWriter{
			ResponseWriter: c.Writer,
			Head:           make(http.Header),
		}
		c.Writer = resWriter

		// 创建访问日志对象
		accessLog := access_service.NewAccessLog(c)

		// 处理请求
		c.Next()

		// 设置响应内容并保存访问日志
		accessLog.SetResponse(resWriter.Body)
		accessLog.Save()
	}
}