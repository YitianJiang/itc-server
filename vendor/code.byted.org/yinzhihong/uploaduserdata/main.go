package uploaduserdata

import "github.com/astaxie/beego/logs"

func test(){
	message := map[string]string{
		"time":"time",
		"action":"action",
		"ip":"ip",
		"username":"username",
		"domain":"domain",
		"ua":"ua",
		"path":"path",
		"title":"title",
		"psm":"psm",
	}
	if err := UploadLog(message); err != nil{
		logs.Error(err.Error())
	}
}