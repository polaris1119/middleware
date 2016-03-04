package middleware

import (
	"net/http"
	"sort"
	"time"

	"github.com/labstack/echo"
	"github.com/polaris1119/goutils"
	"github.com/polaris1119/logger"
	"github.com/polaris1119/nosql"
)

// EchoCache 用于 echo 框架的缓存中间件
func EchoCache() echo.MiddlewareFunc {
	return func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()

			if req.Method == "GET" {

				if len(req.Form) == 0 {
					c.Form("from")
				}

				filter := map[string]bool{
					"from":      true,
					"sign":      true,
					"nonce":     true,
					"timestamp": true,
				}
				var keys = make([]string, 0, len(req.Form))
				for key := range req.Form {
					if _, ok := filter[key]; !ok {
						keys = append(keys, key)
					}
				}

				sort.Sort(sort.StringSlice(keys))

				buffer := goutils.NewBuffer()
				for _, k := range keys {
					buffer.Append(k).Append("=").Append(req.Form.Get(k))
				}

				cacheKey := goutils.Md5(req.Method + req.URL.Path + buffer.String())

				c.Set(nosql.CacheKey, cacheKey)

				value, compressor, ok := nosql.DefaultLRUCache.GetAndUnCompress(cacheKey)
				if ok {
					cacheData, ok := compressor.(*nosql.CacheData)
					if ok {
						logger.Debugln("cache hit:", cacheData.StoreTime, "now:", time.Now())

						// 1分钟更新一次
						if time.Now().Sub(cacheData.StoreTime) >= time.Minute {
							go h(c)
						}

						return c.JSONBlob(http.StatusOK, value)
					}
				}
			}

			if err := h(c); err != nil {
				c.Error(err)
			}

			return nil
		}
	}
}
