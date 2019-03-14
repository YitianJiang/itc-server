package http_util

import (
	"code.byted.org/clientQA/pkg/const"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"errors"
)

type PeopleDepartment struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}
type PeopleUser struct {
	Id         uint             `json:"id"`
	Name       string           `json:"name"`
	Email      string           `json:"email"`
	Department PeopleDepartment `json:"department"`
}
type PeopleRet struct {
	Success   bool         `json:"success"`
	Employees []PeopleUser `json:"employees"`
}

//const peopleAPI = "http://open.byted.org/people/employee/"
const peopleAPI = "https://rocket.bytedance.net/project_manager/users/search"

func GetPeopleSysUser(email string) (error, PeopleRet) {
	logs.Info("%v", email)
	values := map[string]string{"email": email}
	err, body := GetWithBasicAuth(peopleAPI, values, _const.Kani_appid, _const.Kani_apppwd)
	var p PeopleRet
	if err != nil {
		logs.Error("%s", "访问kani失败，无法读取body")
		return errors.New("访问kani失败，无法读取body"), p
	}
	logs.Info("%v", string(body))
	if err := json.Unmarshal(body, &p); err != nil {
		return errors.New("访问kani失败，无法解析"), p
	}
	return nil, p
}
