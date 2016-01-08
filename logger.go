package middleware

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/net/context"

	"github.com/labstack/echo"
	"github.com/polaris1119/logger"
)

// EchoLogger 用于 echo 框架的日志中间件
func EchoLogger() echo.MiddlewareFunc {
	return func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			res := c.Response()

			objLogger := logger.NewLoggerContext(c.Context)
			c.Set("logger", objLogger)

			remoteAddr := req.RemoteAddr
			if ip := req.Header.Get(echo.XRealIP); ip != "" {
				remoteAddr = ip
			} else if ip = req.Header.Get(echo.XForwardedFor); ip != "" {
				remoteAddr = ip
			} else {
				remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
			}

			start := time.Now()
			if err := h(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()
			method := req.Method
			path := req.URL.Path
			if path == "" {
				path = "/"
			}
			size := res.Size()
			code := res.Status()

			uri := fmt.Sprintf("[%s %s %s %d %s %d]", remoteAddr, method, path, code, stop.Sub(start), size)
			objLogger.SetContext(context.WithValue(c.Context, "uri", uri))
			objLogger.Flush()

			return nil
		}
	}
}
