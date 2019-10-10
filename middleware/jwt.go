package middleware

import (
	"fmt"
	"net/http"

	uploadlog "code.byted.org/clientQA/itc-server/upload_log"

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
		if token == "" {
			code = _const.ERROR_AUTH_CHECK_TOKEN_FAIL
		} else {
			claim, flag := ParseTokenString(token)
			if !flag {
				code = _const.ERROR_AUTH_CHECK_TOKEN_FAIL
			} else {
				fmt.Println(claim)
				c.Set("username", claim["name"])
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

		go uploadlog.UploadViaHTTP(c)

		c.Next()
	}
}
