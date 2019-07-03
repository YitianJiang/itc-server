package middleware

import (
	"net/http"

	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/gopkg/logs"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var jwtSecret = []byte("itc_jwt_secret")

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func ParseToken(token string) (*Claims, error) {
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	logs.Info("jwtSecret: ", string(jwtSecret))
	if tokenClaims != nil {
		if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return claims, nil
		}
	}
	return nil, err
}
func ParseTokenString(token string) (jwt.MapClaims, bool) {
	t, _ := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	valid := t.Valid
	claim, _ := t.Claims.(jwt.MapClaims)
	return claim, valid
}
func JWTCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		var data interface{}
		code = _const.SUCCESS
		header := c.Request.Header
		if header == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errorCode": _const.ERROR_AUTH_CHECK_TOKEN_FAIL,
				"message":   _const.GetMsg(code),
				"data":      _const.GetMsg(code),
			})
			c.Abort()
			return
		}
		token := header.Get("Authorization")
		var username string
		if token == "" {
			code = _const.ERROR_AUTH_CHECK_TOKEN_FAIL
		} else {
			claim, flag := ParseTokenString(token)
			if !flag {
				code = _const.ERROR_AUTH_CHECK_TOKEN_FAIL
			} else {
				username = claim["name"].(string)
				c.Set("username", username)
			}
		}
		if code != _const.SUCCESS {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errorCode": code,
				"message":   _const.GetMsg(code),
				"data":      data,
			})
			c.Abort()
			return
		}
		//日志上报
		//uploadMap := map[string]string{
		//	"time":   strconv.FormatInt(time.Now().Unix(), 10),
		//	"action": "enter",
		//	"ip":     c.ClientIP(),
		//	//"username": username,
		//	"username": "yinzhihong",
		//	"domain":   header.Get("Origin"),
		//	"ua":       header.Get("User-Agent"),
		//	"path":     strings.Trim(header.Get("Referer"), header.Get("Origin")),
		//	"title":    "预审平台",
		//	"psm":      "toutiao.clientqa.itcserver",
		//}
		//if err := uploaduserdata.UploadLog(uploadMap); err != nil {
		//	logs.Error(err.Error())
		//	//c.Abort()
		//	//return
		//}
		c.Next()
	}
}
