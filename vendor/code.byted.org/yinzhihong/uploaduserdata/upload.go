package uploaduserdata

import (
	"code.byted.org/gopkg/logs"
	"time"
)

/**
example：oa.log
oa日志格式:'[$time_local]\t"$request"\t$status\t$body_bytes_sent\t"$http_referer"'
                            '\t"$http_user_agent"\t$content_length\t$request_time\t$upstream_response_time\t$sent_http_content_type'
                            '\t$sent_http_transfer_encoding\t$upstream_addr\t$http_x_forwarded_for\t$remote_addr\t$host'
*/

//业务方编写callback函数解析并按照要求的格式返回上报数据

var callback = func(message map[string]string)(*Msg, error){
	var err error
	msgJson := Msg{}
	msgJson.Event.Time = message["time"]
	msgJson.Event.Action = message["action"]
	msgJson.User.Ip = message["ip"]
	msgJson.User.User_name = message["username"]
	msgJson.Header.Domain = message["domain"]
	msgJson.Header.Ua = message["ua"]
	msgJson.Header.Path = message["path"]
	msgJson.Header.Title = message["title"]
	msgJson.Source = "rocket"
	return &msgJson, err
}
func UploadLog(message map[string]string) error{
	psm := message["psm"]
	delete(message, "psm")
	if err := Publish(callback, message, psm, time.Now().Unix()); err != nil{
		logs.Error("上报用户行为错误!, ", err)
		return err
	}
	return nil
}
