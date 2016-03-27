package middleware

import (
	"fmt"
	"net"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/polaris1119/logger"
	"github.com/twinj/uuid"
	"golang.org/x/net/context"
)

const HeaderKey = "X-Request-Id"

// EchoLogger 用于 echo 框架的日志中间件
func EchoLogger() echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) error {
			start := time.Now()

			req := c.Request().(*standard.Request).Request
			resp := c.Response().(*standard.Response)

			objLogger := logger.NewLoggerContext(c.NetContext())

			// 用 req.ParseForm 会导致数据丢失，原因未知
			if len(req.Form) == 0 {
				c.FormValue("from")
			}
			objLogger.Infoln("input params:", req.Form)

			if c.NetContext() == nil {
				c.SetNetContext(context.WithValue(context.Background(), "logger", objLogger))
			}

			remoteAddr := req.RemoteAddr
			if ip := req.Header.Get(echo.XRealIP); ip != "" {
				remoteAddr = ip
			} else if ip = req.Header.Get(echo.XForwardedFor); ip != "" {
				remoteAddr = ip
			} else {
				remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
			}

			id := func(c echo.Context) string {

				id := req.Header.Get(HeaderKey)
				if id == "" {
					id = c.FormValue("request_id")
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
				objLogger.SetContext(context.WithValue(c.NetContext(), "uri", uri))
				objLogger.Flush()
			}()

			if err := next.Handle(c); err != nil {
				return err
			}
			return nil
		})
	}
}
