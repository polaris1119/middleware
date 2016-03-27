package middleware

import (
	"net/http"
	"net/url"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
)

type Signature func(url.Values, string) string

// EchoAuth 用于 echo 框架的签名校验中间件
func EchoAuth(signature Signature, secretKey string) echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(ctx echo.Context) error {
			req := ctx.Request().(*standard.Request).Request

			if len(req.Form) == 0 {
				ctx.Form("from")
			}

			if sign := signature(req.Form, secretKey); sign != ctx.Form("sign") {
				return ctx.String(http.StatusBadRequest, `400 Bad Request`)
			}

			if err := next.Handle(ctx); err != nil {
				return err
			}

			return nil
		})
	}
}
