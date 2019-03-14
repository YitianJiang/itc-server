package dal

import "code.byted.org/gopkg/gorm"

type BinaryDetectTool struct {
	gorm.Model
	name string 		`json:"name"`
	desc string 		`json:"desc"`
	platform int 		`json:"platform"`
}
