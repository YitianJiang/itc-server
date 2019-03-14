package throttle

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	DEFAULT_QPS_LIMIT = 50000
	DEFAULT_MAX_CON   = 10000
)

func Throttle() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !globalLimiter.TakeCon() {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		defer func() {
			e := recover()
			globalLimiter.ReleaseCon()
			if e != nil {
				panic(e)
			}
		}()

		if !globalLimiter.TakeQPS() {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
	}
}
