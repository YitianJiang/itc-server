package dal

import (
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

type BinaryDetectTool struct {
	gorm.Model
	Name string 		`json:"name"`
	Desc string 		`json:"desc"`
	Platform int 		`json:"platform"`
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
	if err := connection.Table(BinaryDetectTool{}.TableName()).Where(condition).Find(&detect).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	return &detect
}
//insert data
func InsertBinaryTool(binaryTool BinaryDetectTool) bool {
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return false
	}
	defer connection.Close()
	if err := connection.Table(BinaryDetectTool{}.TableName()).Create(binaryTool).Error; err != nil {
		logs.Error("insert binary tool failed, %v", err)
		return false
	}
	return true
}
