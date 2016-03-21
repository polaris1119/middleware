package middleware

import (
	"net/http"
	"net/url"
	"sort"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/polaris1119/config"
	"github.com/polaris1119/goutils"
)

// EchoAuth 用于 echo 框架的签名校验中间件
func EchoAuth() echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(ctx echo.Context) error {
			req := ctx.Request().(*standard.Request).Request

			if len(req.Form) == 0 {
				ctx.Form("from")
			}

			if sign := genSign(req.Form); sign != ctx.Form("sign") {
				return ctx.String(http.StatusBadRequest, `400 Bad Request`)
			}

			if err := next.Handle(ctx); err != nil {
				return err
			}

			return nil
		})
	}
}

func genSign(args url.Values) string {
	keys := make([]string, 0, len(args))
	for k := range args {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Sort(sort.StringSlice(keys))

	buffer := goutils.NewBuffer()
	for _, k := range keys {
		buffer.Append(k).Append("=").Append(args.Get(k))
	}

	buffer.Append(config.ConfigFile.MustValue("security", "salt_secret", ""))

	return goutils.Md5(buffer.String())
}
