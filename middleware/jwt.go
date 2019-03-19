package middleware

import (
	"code.byted.org/clientQA/itc-server/const"
	"code.byted.org/gopkg/logs"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
)
var jwtSecret = []byte("itc_jwt_secret")
type Claims struct {
	/*Username string `json:"itc"`
	Password string `json:"itc"`*/
	jwt.StandardClaims
}

/*unc GenerateToken(username, password string) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(3 * time.Hour)
	claims := Claims{
		username,
		password,
		jwt.StandardClaims {
			ExpiresAt : expireTime.Unix(),
			Issuer : "gin-blog",
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)
	return token, err
}*/

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
func JWTCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		var data interface{}
		code = _const.SUCCESS
		token := c.Query("token")
		if token == "" {
			code = _const.INVALID_PARAMS
		} else {
			_, err := ParseToken(token)
			if err != nil {
				switch err.(*jwt.ValidationError).Errors {
				case jwt.ValidationErrorExpired:
					code = _const.ERROR_AUTH_CHECK_TOKEN_TIMEOUT
				default:
					code = _const.ERROR_AUTH_CHECK_TOKEN_FAIL
				}
			}
		}
		if code != _const.SUCCESS {
			c.JSON(http.StatusUnauthorized, gin.H{
				"errorCode": code,
				"message":  _const.GetMsg(code),
				"data": data,
			})

			c.Abort()
			return
		}
		c.Next()
	}
}