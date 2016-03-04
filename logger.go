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
			start := time.Now()

			req := c.Request()
			resp := c.Response()

			objLogger := logger.NewLoggerContext(c.Context)

			// 用 req.ParseForm 会导致数据丢失，原因未知
			if len(req.Form) == 0 {
				c.Form("from")
			}
			objLogger.Infoln("input params:", req.Form)

			if c.Context == nil {
				c.Context = context.WithValue(context.Background(), "logger", objLogger)
			}

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

			resp.Header().Set(HeaderKey, id)

			defer func() {
				method := req.Method
				path := req.URL.Path
				if path == "" {
					path = "/"
				}
				size := resp.Size()
				code := resp.Status()

				stop := time.Now()
				// [remoteAddr method path request_id code time size]
				uri := fmt.Sprintf("[%s %s %s %s %d %s %d]", remoteAddr, method, path, id, code, stop.Sub(start), size)
				objLogger.SetContext(context.WithValue(c.Context, "uri", uri))
				objLogger.Flush()
			}()

			if err := h(c); err != nil {
				c.Error(err)
			}

			return nil
		}
	}
}
