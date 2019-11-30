package detect

type confirmParams struct {
	TaskID uint   `json:"taskId"`
	ToolID int    `json:"toolId"`
	Status int    `json:"status"`
	Remark string `json:"remark"`

	// Only used in Android.
	// ID is the id of table tb_detect_content_detail if the type is method or string.
	// ID is the array index of table tb_perm_app_relation's field perm_infos
	// if the type is permission. And must use ID-1 because it starts from 1.
	ID int `json:"id"`
	// Index corresponds to the table tb_perm_app_relation's field perm_infos
	// because .aab package may have more than one record in single task.
	Index int `json:"index"`
	// 0-->method/string 1-->permission
	TypeAndroid int `json:"type"`

	// Only used in iOS.
	// 1-->blacklist(string) 2-->method 3-->privacy(permission)
	TypeiOS int `json:"confirmType"`
	// Name=methodName+className if the type is method.
	Name string `json:"confirmContent"`

	// The field will be filled in the code.
	APPID      string
	Platform   int
	APPVersion string
	Item       *Item
	Confirmer  string
}
