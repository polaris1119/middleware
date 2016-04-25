package middleware

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/polaris1119/goutils"
)

// EchoAsync 用于 echo 框架的异步处理中间件
func EchoAsync() echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(ctx echo.Context) error {
			req := ctx.Request().(*standard.Request).Request

			if req.Method != "GET" {
				// 是否异步执行
				async := goutils.MustBool(ctx.FormValue("async"), false)
				if async {
					go next.Handle(ctx)

					result := map[string]interface{}{
						"code": 0,
						"msg":  "ok",
						"data": nil,
					}
					return ctx.JSON(http.StatusOK, result)
				}
			}

			if err := next.Handle(ctx); err != nil {
				return err
			}

			return nil
		})
	}
}