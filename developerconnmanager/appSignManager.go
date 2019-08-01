package developerconnmanager

import (
	_const "code.byted.org/clientQA/itc-server/const"
	devconnmanager "code.byted.org/clientQA/itc-server/database/dal/AppleConnMannagerModel"
	"code.byted.org/clientQA/itc-server/utils"
	"github.com/gin-gonic/gin"
	"strings"
)

/**
API 3-1：根据业务线appid，返回app相关list
*/
func GetAppDetailInfo(c *gin.Context) {
	username, ok := c.GetQuery("user_name")
	if !ok {
		utils.AssembleJsonResponse(c, -1, "缺少user_name参数", "")
		return
	}
	app_id, ok := c.GetQuery("app_id")
	if !ok {
		utils.AssembleJsonResponse(c, -1, "缺少app_id参数", "")
		return
	}
	app_acc_certs := devconnmanager.QueryAppAccountCert(map[string]interface{}{
		"app_id": app_id,
	})
	if app_acc_certs == nil {
		utils.AssembleJsonResponse(c, -2, "数据库查询tt_app_account_cert失败！", "")
		return
	} else if len(*app_acc_certs) == 0 {
		utils.AssembleJsonResponse(c, -3, "未查询到该app_id下的账号信息！", "")
	}

	var fQueryResult []devconnmanager.APPandCert
	sql := "select aac.app_name,aac.app_type,aac.id as app_acount_id,aac.team_id,aac.account_verify_status,aac.account_verify_user," +
		"ac.cert_id,ac.cert_type,ac.cert_expire_date,ac.cert_download_url,ac.priv_key_url from tt_app_account_cert acc, tt_apple_certificate ac" +
		" where acc.app_id = '" + app_id + "' and aac.deleted IS NULL and (aac.dev_cert_id = ac.id or aac.dist_cert_id = ac.id) and ac.deleted_at IS NULL "
	f_query := devconnmanager.QueryWithSql(sql, &fQueryResult)

	var resourcPerm devconnmanager.GetPermsResponse
	url := _const.Certain_Resource_All_PERMS_URL + "employeeKey=" + username + "&resourceKeys[]=" + ""
	result := QueryPerms(url, &resourcPerm)

}

func GetResourcePermType(c *gin.Context, teamIds []string, username string) (map[string]int, bool) {
	url := _const.Certain_Resource_All_PERMS_URL + "employeeKey=" + username
	var resourceMap = make(map[string]string)
	for _, teamId := range teamIds {
		lowTeamId := strings.ToLower(teamId)
		resource := lowTeamId + "_space_account"
		resourceMap[resource] = teamId
		url += "&resourceKeys[]=" + resource
	}
	var resourcPerm devconnmanager.GetPermsResponse
	result := QueryPerms(url, &resourcPerm)
	if !result || resourcPerm.Errno != 0 {
		utils.AssembleJsonResponse(c, -4, "查询权限失败！", "")
		return nil, false
	}
	for k := range resourcPerm.Data {

	}

}
