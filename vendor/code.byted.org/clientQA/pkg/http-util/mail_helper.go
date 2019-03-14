package http_util

import (
	"github.com/gin-gonic/gin"
	"gopkg.in/cas.v2"
	"net/http"
	"code.byted.org/gopkg/logs"
)

const Sendemailapi = "https://fatal.bytedance.net/mail_util/mail_send/send"

func SendMail(c *gin.Context) {
	username := cas.Username(c.Request)
	if username == "" {
		username = "wangshiyu"
	}
	logs.Debug("%v",username)
	//是否要加一个可以发邮件权限的校验？
	//db := c.MustGet("DB").(*gorm.DB)
	//if !isPermissioned(db,curuser.ID,permission_can_send_version_mail){
	//	SetNoPermissionRet(c)
	//	return
	//}
	mail := c.DefaultPostForm("mail","")
	logs.Debug(mail)
	_, body := PostJsonHttp(Sendemailapi, []byte(mail))
	c.JSON(http.StatusOK, gin.H{"message":"success", "errorCode":0,"data":string(body)})
}