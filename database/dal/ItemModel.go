package dal

import "code.byted.org/gopkg/gorm"

type ItemStruct struct {
	gorm.Model
	QuestionType int
	KeyWord int
	FixWay int
	CheckContent string
	Resolution string
	Regulation string
	RegulationUrl string
}
func (ItemStruct) TableName() string {
	return "tb_item"
}
