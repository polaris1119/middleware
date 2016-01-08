package middleware

import (
	"fmt"
	"net"
	"time"

	"golang.org/x/net/context"

	"github.com/labstack/echo"
	"github.com/polaris1119/logger"
	"github.com/twinj/uuid"
)

const HeaderKey = "X-Request-Id"

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

			id := func(c *echo.Context) string {

				id := req.Header.Get(HeaderKey)
				if id == "" {
					id = c.Query("request_id")
					if id == "" {
						id = uuid.NewV4().String()
					}
				}

				c.Set("request_id", id)

				return id
			}(c)

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

			// [remoteAddr method path request_id code time size]
			uri := fmt.Sprintf("[%s %s %s %s %d %s %d]", remoteAddr, method, path, id, code, stop.Sub(start), size)
			objLogger.SetContext(context.WithValue(c.Context, "uri", uri))
			objLogger.Flush()

			c.Response().Header().Set(HeaderKey, id)

			return nil
		}
	}
}
