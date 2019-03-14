package dal

import (
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

type BinaryDetectTool struct {
	gorm.Model
	name string 		`json:"name"`
	desc string 		`json:"desc"`
	platform int 		`json:"platform"`
}
func (BinaryDetectTool) TableName() string {
	return "tb_binary_detect_tool"
}
//query by map
func QueryBinaryToolsByCondition(condition string) *[]BinaryDetectTool{
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return nil
	}
	defer connection.Close()
	var detect []BinaryDetectTool
	if err := connection.Where(condition).Find(&detect).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	return &detect
}
