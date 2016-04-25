package middleware

import (
	"net/http"
	"sort"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/polaris1119/goutils"
	"github.com/polaris1119/logger"
	"github.com/polaris1119/nosql"
)

type CacheKeyAlgorithm interface {
	GenCacheKey(echo.Context) string
}

type CacheKeyFunc func(echo.Context) string

func (self CacheKeyFunc) GenCacheKey(ctx echo.Context) string {
	return self(ctx)
}

var CacheKeyAlgorithmMap = make(map[string]CacheKeyAlgorithm)

var LruCache = nosql.DefaultLRUCache

// EchoCache 用于 echo 框架的缓存中间件。支持自定义 cache 数量
func EchoCache(cacheMaxEntryNum ...int) echo.MiddlewareFunc {

	if len(cacheMaxEntryNum) > 0 {
		LruCache = nosql.NewLRUCache(cacheMaxEntryNum[0])
	}

	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(ctx echo.Context) error {
			req := ctx.Request().(*standard.Request).Request

			if req.Method == "GET" {
				cacheKey := getCacheKey(req, ctx)

				if cacheKey != "" {
					ctx.Set(nosql.CacheKey, cacheKey)

					value, compressor, ok := LruCache.GetAndUnCompress(cacheKey)
					if ok {
						cacheData, ok := compressor.(*nosql.CacheData)
						if ok {

							// 1分钟更新一次
							if time.Now().Sub(cacheData.StoreTime) >= time.Minute {
								// TODO:雪崩问题处理
								goto NEXT
							}

							logger.Debugln("cache hit:", cacheData.StoreTime, "now:", time.Now())
							return ctx.JSONBlob(http.StatusOK, value)
						}
					}
				}
			}

		NEXT:
			if err := next.Handle(ctx); err != nil {
				return err
			}

			return nil
		})
	}
}

func getCacheKey(req *http.Request, ctx echo.Context) string {
	if len(req.Form) == 0 {
		ctx.FormValue("from")
	}

	cacheKey := ""
	if cacheKeyAlgorithm, ok := CacheKeyAlgorithmMap[ctx.Path()]; ok {
		// nil 表示不缓存
		if cacheKeyAlgorithm != nil {
			cacheKey = cacheKeyAlgorithm.GenCacheKey(ctx)
		}
	} else {
		cacheKey = defaultCacheKeyAlgorithm(req)
	}

	return cacheKey
}

func defaultCacheKeyAlgorithm(req *http.Request) string {
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

	return goutils.Md5(req.Method + req.URL.Path + buffer.String())
}
