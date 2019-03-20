package main

import (
	"code.byted.org/clientQA/itc-server/conf"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/pkg/http-util"
	"code.byted.org/clientQA/pkg/request-processor/request-dal"
	"code.byted.org/gin/ginex"
	"code.byted.org/gopkg/logs"
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/cas.v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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
	InitRouter(r)
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
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar")      // 跨域关键设置 让浏览器可以解析
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
//sso login
func casInitAndVerify() gin.HandlerFunc {

	u, _ := url.Parse("https://sso.bytedance.com/cas/")
	client := cas.NewClient(&cas.Options{URL: u})
	handler := client.HandleFunc(func(w http.ResponseWriter, r *http.Request) {
		userName := cas.Username(r)
		attr := cas.Attributes(r)
		full_name := attr.Get("full_name")
		logs.Debug("%s", userName)
		if full_name == "" {
			logs.Debug("%v", r.Header.Get("origin"))
			if strings.Contains(r.Header.Get("origin"), "cloud.bytedance.net") ||
				strings.Contains(r.Header.Get("origin"), "sandbox.cloud.byted.org") ||
				strings.Contains(r.Header.Get("origin"), "localhost") ||
				strings.Contains(r.Header.Get("origin"), "10.2.219.30") {
				if r.Method == "GET" || r.Method == "POST" {
					w.WriteHeader(http.StatusUnauthorized)
				}
			}
			return
		}
		user := request_dal.GetUserInfo(userName)
		if user == nil {
			strs := strings.Split(full_name, " ")
			if len(strs) > 1 {
				full_name = strs[0]
				num, _ := strconv.Atoi(strs[1])
				var user request_dal.Struct_User
				user.Name = userName
				user.Email = userName + "@bytedance.com"
				user.Employeenumber = uint(num)
				user.Full_name = full_name
				user.PermissionLevel = "5"
				request_dal.CreateUser(user)
			} else {
				userstruct := http_util.GetUserAvatar(userName + "@bytedance.com")
				var user request_dal.Struct_User
				user.Name = userName
				user.Email = userName + "@bytedance.com"
				user.Employeenumber = userstruct.Employeenumber
				user.Full_name = full_name
				user.PermissionLevel = "5"
				request_dal.CreateUser(user)
			}
		} else {
			if user.Full_name == "" {
				request_dal.UpdateUserInfo(full_name, user.ID)
			}
		}
	})

	return func(c *gin.Context) {
		c.Set("casClient", client)
		logs.Debug("%s", c.Request.URL)
		handler.ServeHTTP(c.Writer, c.Request)

		if c.Request.URL.Path == "/accounts/login" {
			next := c.Query("next")
			if next != "" {
				c.Abort()
				rCookie, err := c.Request.Cookie("_cas_session")
				if err == nil {
					logs.Debug("%s", "setcookie:"+rCookie.Value)
					c.Writer.Header().Set("Set-Cookie", "_cas_session="+rCookie.Value+"; Max-Age=86400; path=/")
				}
				c.Redirect(http.StatusMovedPermanently, next)
				return
			}
		}
		if c.Request.Method == "OPTIONS" {
			c.Abort()
			c.JSON(http.StatusOK, gin.H{"message": "success", "errorCode": 0, "data": ""})
			return
		}
		if c.Writer.Status() == http.StatusUnauthorized {
			c.Abort()
			c.Writer.Header().Del("Set-Cookie")
			c.JSON(http.StatusUnauthorized, gin.H{"message": "error", "errorCode": 1, "data": ""})
			return
		}
		c.Next()
	}
}
