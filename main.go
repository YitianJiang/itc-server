package main

import (
	"fmt"
	"net/http"
	"strings"

	"code.byted.org/clientQA/itc-server/conf"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/detect"
	"code.byted.org/gin/ginex"
	"github.com/gin-gonic/gin"
)

func main() {
	ginex.Init()
	r := ginex.Default()
	r.GET("/ping", Ping)
	// add your handlers here
	//init
	conf.InitConfiguration()
	//r.Use(casInitAndVerify())
	r.Use(Cors())
	database.InitDB()
	if err := database.InitDBHandler(); err != nil {
		panic(fmt.Sprintf("failed to initialize the global database handler: %v", err))
	}

	InitRouter(r)
	detect.InitCron()
	r.Run()
}

// 跨域
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		//请求方法
		method := c.Request.Method
		//请求头部
		origin := c.Request.Header.Get("Origin")
		//声明请求头keys
		var headerKeys []string
		for k, _ := range c.Request.Header {
			headerKeys = append(headerKeys, k)
		}
		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		} else {
			headerStr = "access-control-allow-origin, access-control-allow-headers"
		}
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			// 这是允许访问所有域
			c.Header("Access-Control-Allow-Origin", origin)
			//服务器支持的所有跨域请求的方法,为了避免浏览次请求的多次'预检'请求
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			//header的类型
			c.Header("Access-Control-Allow-Headers", " Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With,X-Nt-Engine")
			//允许跨域设置，可以返回其他字段
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar") // 跨域关键设置 让浏览器可以解析
			//缓存请求信息 单位为秒
			c.Header("Access-Control-Max-Age", "172800")
			//跨域请求是否需要带cookie信息 默认设置为true
			c.Header("Access-Control-Allow-Credentials", "true")
			//设置返回格式是json
			c.Set("content-type", "application/json")
		}

		//放行所有OPTIONS方法
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}
		// 处理请求
		c.Next()
	}
}
