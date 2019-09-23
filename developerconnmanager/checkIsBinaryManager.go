package developerconnmanager

import (
	"bytes"
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/gopkg/logs"
	"code.byted.org/yuyilei/bot-api/form"
	"code.byted.org/yuyilei/bot-api/service"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"code.byted.org/clientQA/itc-server/utils"
	"io/ioutil"
	"net/http"
)

type CheckInBinaryStatus struct {
	OperateUser   string             `json:"operate_user"    binding:"required"`
	RepoName      string			 `json:"repo_name"       binding:"required"`
	Version       string 	         `json:"version"         binding:"required"`
	BuildResult   int                `json:"build_result"    binding:"required"`
	BuildFrom     BuildFromObj       `json:"build_from"      binding:"required"`
	IosExtInfo    IosExtInfoObj      `json:"ios_ext_info"`
}

type BuildFromObj struct {
	Type          string             `json:"type"            binding:"required"`
}

type IosExtInfoObj struct {
	IsBinary      bool               `json:"is_binary"`
}

type UserInfoReqToLark struct {
	Email         string             `json:"email"           binding:"required"`
}

type UserInfoGetFromLark struct {
	Data          UserIdInfoFromLark `json:"data"            binding:"required"`
}

type UserIdInfoFromLark struct {
	OpenId        string	         `json:"open_id"         binding:"required"`
	UserId        string             `json:"user_id"         binding:"required"`
}

func CheckIsBinary(c *gin.Context)  {
	logs.Info("检测cony合代码，对应库是不是二进制成功")
	var body CheckInBinaryStatus
	err := c.ShouldBindJSON(&body)
	utils.RecordError("参数绑定失败", err)
	if err != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	if body.BuildFrom.Type == "cony" && body.IosExtInfo.IsBinary && body.BuildResult == 3{
		botService := service.BotService{}
		botService.SetAppIdAndAppSecret(_const.BotApiId,_const.BotAppSecret)
		larkIdRes,larkUserId := GetUserIdFromLark(body.OperateUser)
		if !larkIdRes{
			c.JSON(http.StatusInternalServerError, gin.H{
				"message":   "获取AT人的id错误",
				"error_code": "1",
				"data": "",
			})
			return
		}
		chatId := "oc_d96be0fa3736fbfa1364bc57f329eeef"
		openIds := []string{larkUserId.Data.OpenId}
		_, err = botService.InviteUser2Chat(chatId,openIds)
		utils.RecordError("bot发送消息错误", err)
		contentText := form.SendMessageForm{}
		msgType := "text"
		content := fmt.Sprintf("%s库的%s版本二进制失败了，主干分支问题请第一时间跟进<at user_id=\"%s\">test</at>",body.RepoName,body.Version,larkUserId.Data.UserId)
		contentText.OpenChatID = &chatId
		contentText.MsgType = &msgType
		contentText.Content.Text = &content
		_, errorSendInfo := botService.SendMessage(contentText)
		utils.RecordError("bot发送消息错误", errorSendInfo)
		if errorSendInfo != nil{
			c.JSON(http.StatusInternalServerError, gin.H{
				"message":   "发送二进制报警消息失败",
				"error_code": "2",
				"data": "",
			})
			return
		}else {
			c.JSON(http.StatusOK, gin.H{
				"message":   "发送二进制报警消息成功",
				"error_code": "",
				"data": "success",
			})
		}
	}else {
		c.JSON(http.StatusOK, gin.H{
			"message":   "非cony合码，或者非二进制都不需要监控",
			"error_code": "",
			"data": "success",
		})
	}

}


func GetUserIdFromLark(userEmail string) (bool,*UserInfoGetFromLark){
	tokenFormservice,err := service.GetTenantAccessToken(_const.BotApiId,_const.BotAppSecret)
	utils.RecordError("获取TenantAccessToken失败", err)
	var reqUserToLark UserInfoReqToLark
	reqUserToLark.Email = userEmail+"@bytedance.com"
	logs.Info(tokenFormservice.TenantAccessToken)
	var resUserInfo UserInfoGetFromLark
	reqResult := PostToLarkGetInfo("POST","https://open.feishu.cn/open-apis/user/v4/email2id","Bearer "+tokenFormservice.TenantAccessToken,&reqUserToLark,&resUserInfo)
	if reqResult{
		return true,&resUserInfo
	}else{
		return false,&resUserInfo
	}
}

func PostToLarkGetInfo(method, url, tokenString string, objReq, objRes interface{}) bool {
	var rbodyByte *bytes.Reader
	if objReq != nil {
		bodyByte, _ := json.Marshal(objReq)
		logs.Info(string(bodyByte))
		rbodyByte = bytes.NewReader(bodyByte)
	} else {
		rbodyByte = nil
	}
	client := &http.Client{}
	var err error
	var request *http.Request
	if rbodyByte != nil {
		request, err = http.NewRequest(method, url, rbodyByte)
	} else {
		request, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		logs.Info("新建request对象失败")
		return false
	}
	request.Header.Set("Authorization", tokenString)
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		logs.Info("发送请求失败")
		return false
	}
	defer response.Body.Close()
	logs.Info("状态码：%d", response.StatusCode)
	if !AssertResStatusCodeOK(response.StatusCode) {
		logs.Info("查看返回状态码")
		logs.Info(string(response.StatusCode))
		responseByte, _ := ioutil.ReadAll(response.Body)
		logs.Info("苹果失败返回response\n：%s", string(responseByte))
		return false
	} else {
		responseByte, err := ioutil.ReadAll(response.Body)
		logs.Info("查看苹果的返回值")
		logs.Info(string(responseByte))
		if err != nil {
			logs.Info("读取respose的body内容失败")
			return false
		}
		logs.Info("苹果成功返回response\n：%s", string(responseByte))
		json.Unmarshal(responseByte, objRes)
		return true
	}
}
