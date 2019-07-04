package uploaduserdata
type Msg struct {
	Event  EventMsg  `json:"event"`
	Header HeaderMsg `json:"header"`
	User   UserMsg   `json:"user"`
	Source string    `json:"source"`
}
type EventMsg struct { //  ***Event必填
	Action string `json:"action"` //事件具体动作
	Time   string `json:"time"`   //事件发生时间
}
type HeaderMsg struct { //  ***Header必填
	Domain string `json:"domain"` //domain
	Title  string `json:"title"`  //页面名称
	Ua     string `json:"ua"`     //User-agent
	Path   string `json:"path"`   //url去除domain部分
}
type UserMsg struct { //  ***IP必填其他选填
	User_name  string     `json:"user_name"`  //邮箱前缀
	User_id    string     `json:"user_id"`    //工号
	Ip         string     `json:"ip"`         //员工IP
	Department Department `json:"department"` //员工部门信息
}
type Department struct {
	Dep_id   string `json:"dep_id"`   //部门ID
	Dep_name string `json:"dep_name"` //部门名称
}