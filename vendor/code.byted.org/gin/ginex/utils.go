package ginex

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func GetClientIP(ctx *gin.Context) string {
	// According to 王宁's email:《Fwd: 动态加速可能需要业务逻辑做些调整》
	// Ali-CDN-Real-IP has the highest priority
	clientIP := ctx.GetHeader("Ali-CDN-Real-IP")
	if len(clientIP) != 0 {
		return clientIP
	}
	clientIP = ctx.GetHeader("X-Real-IP")
	if len(clientIP) != 0 {
		return clientIP
	}
	ips := ctx.GetHeader("X-Forwarded-For")
	if len(ips) != 0 {
		clientIP = strings.Split(ips, ",")[0]
		return clientIP
	}
	// try to call gin.Context.ClientIP after all
	return ctx.ClientIP()
}
