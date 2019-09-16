package dal

import (
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/logs"
	"time"
)

type DetectAnaStruct struct {
	Total 			int				`json:"total"`
	AppId 			int				`json:"appId"`
	Platform		int				`json:"platform"`
}

type QueryReportInfo struct {
	AppId 			int 			`json:"appId"`
	Platform        int 			`json:"platform"`
	StartTime 		time.Time		`json:"startTime"`
	EndTime			time.Time		`json:"endTime"`
}
type ChartInfoStruct struct {
	Type 			string			`json:"type"`
	Title			string			`json:"title"`
	Total 			string			`json:"total"`

}

func QueryReportsInfo (condition string) (*[]DetectStruct,error) {
	connection,err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil,err

	}
	defer connection.Close()
	db := connection.Table(DetectStruct{}.TableName()).LogMode(_const.DB_LOG_MODE)
	var result []DetectStruct
	if err := db.Where(condition).Find(&result).Error;err != nil {
		logs.Error("query reports info failed, infos:"+condition+",errInfo:%v",err)
		return nil,err
	}
	return &result,nil
}