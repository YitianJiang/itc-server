package developerconnmanager

import (
	"bytes"
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/context"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/tos"
	"code.byted.org/yuyilei/bot-api/form"
	"code.byted.org/yuyilei/bot-api/service"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type ReqTFReviewInfoFromClient struct {
	VersionLong               string            		`form:"version_long"		    binding:"required"`
	AppId                     string					`form:"app_id"                  binding:"required"`
	VersionShort              string 	        		`form:"version_short"           binding:"required"`
}

//苹果后台build信息和review信息返回值解析model *************Start*************
type AttributeItemInfo struct {
	LongVersion               string            		`json:"version"                 binding:"required"`
	ProcessingState   		  string					`json:"processingState"         binding:"required"`
}

type MainDataListItem struct {
	Type 					  string			        `json:"type"                    binding:"required"`
	BuildId					  string                    `json:"id"                      binding:"required"`
	Attributes                AttributeItemInfo	        `json:"attributes"              binding:"required"`
}

type IncludedAttributeItemInfo struct {
	BetaReviewState           string			        `json:"betaReviewState"         binding:"required"`
}

type IncludeDataListItem struct {
	Type 					  string			        `json:"type"                    binding:"required"`
	BuildId					  string                    `json:"id"                      binding:"required"`
	ReviewAttributes          IncludedAttributeItemInfo	`json:"attributes"              binding:"required"`
}

type ResTFReviewInfoFromApple struct {
	Data  					  []MainDataListItem        `json:"data"                    binding:"required"`
	Included                  []IncludeDataListItem		`json:"included"                binding:"required"`
}
//苹果后台build信息和review信息返回值解析model *************End*************
//苹果后台TF 提审的req、res解析model *************Start*************
type BuildSubmInfoItem struct {
	BuildId					  string                    `json:"id"                      binding:"required"`
	BuildType 				  string			        `json:"type"                    binding:"required"`
}
type BuildSection struct {
	DataZone 				  BuildSubmInfoItem			`json:"data"                    binding:"required"`
}
type RelationSection struct {
	BuildZone             	  BuildSection              `json:"build"                   binding:"required"`
}
type BuildSubmDataSec struct {
	Type					  string			        `json:"type"                    binding:"required"`
	Relationships			  RelationSection           `json:"relationships"           binding:"required"`
}
type ReqSubmData struct {
	Data 					  BuildSubmDataSec          `json:"data"                    binding:"required"`
}

type ResSubmAttrItem struct {
	BetaReviewState           string					`json:"betaReviewState"         binding:"required"`
}
type ResSubmAttrSec struct {
	Attributes 				  ResSubmAttrItem           `json:"attributes"              binding:"required"`
}
type ResSubmData struct {
	Data 					  ResSubmAttrSec            `json:"data"                    binding:"required"`
}
//苹果后台TF 提审的req、res解析model *************End*************

//返回给发布平台的app状态还有build id *************Start*************
type BuildIdOrStatus struct {
	VersionStatus    		  string					`json:"versionStatus"           binding:"required"`
	AppleBuildId			  string					`json:"appleBuildId"            binding:"required"`
}
//返回给发布平台的app状态还有build id *************End*************

//发布平台请求新建Group的参数解析 *************Start*************
type ReqTFLinkFromClient struct {
	AppBuildId				  string                    `json:"app_build_id"            binding:"required"`
	GroupName				  string					`json:"group_name"              binding:"required"`
	AppId                     string					`json:"app_id"                  binding:"required"`
}
//发布平台请求新建Group的参数解析 *************End*************

//请求apple 生产group的model解析 *************Start*************
type AppDataItem struct {
	AppleAppId       	      string                    `json:"id"                      binding:"required"`
	AppleAppType              string                    `json:"type"                    binding:"required"`
}
type AppSection struct {
	Data                      AppDataItem               `json:"data"                    binding:"required"`
}
type RelationShipSection struct{
	AppZone					  AppSection				`json:"app"                     binding:"required"`
}
type AttrItem struct {
	GroupName                 string                    `json:"name"                    binding:"required"`
}
type ReqGroupCreateDataItem struct {
	Type                      string                    `json:"type"                    binding:"required"`
	Relationships			  RelationShipSection       `json:"relationships"           binding:"required"`
	Attributes				  AttrItem					`json:"attributes"              binding:"required"`
}
type ReqGroupCreateData struct {
	Data                      ReqGroupCreateDataItem    `json:"data"                    binding:"required"`
}
//请求值 apple 生产group的model解析 *************End*************

//返回值 apple生成完group后的model解析 *************Start*************
type GroupAttrItem struct {
	GroupName                 string                    `json:"name"                    binding:"required"`
	PublicLinkEnabled		  bool                      `json:"publicLinkEnabled"       binding:"required"`
	PublicLink			      string                    `json:"publicLink"              binding:"required"`
}
type ResCreateGroupDataItem struct {
	Type                      string                    `json:"type"                    binding:"required"`
	Id						  string                    `json:"id"                      binding:"required"`
	Attributes				  GroupAttrItem             `json:"attributes"              binding:"required"`
}
type ResCreateGroupData struct {
	Data                      ResCreateGroupDataItem    `json:"data"                    binding:"required"`
}
//apple生成完group后的model解析 *************End*************

//请求apple 把版本加入Group的请求解析 *************Start*************
type ReqVersionToGroupItem struct {
	Type                      string                    `json:"type"                    binding:"required"`
	Id						  string                    `json:"id"                      binding:"required"`
}
type ReqVersionToGroupData struct {
	Data                      []ReqVersionToGroupItem   `json:"data"                    binding:"required"`
}
//请求apple 把版本加入Group的请求解析 *************End*************

//patch请求，打开apple公链   *************Start*************
type PatchGroupAttrItem struct {
	PublicLinkEnabled		  bool                      `json:"publicLinkEnabled"       binding:"required"`
}
type PatchResCreateGroupDataItem struct {
	Type                      string                    `json:"type"                    binding:"required"`
	Id						  string                    `json:"id"                      binding:"required"`
	Attributes				  PatchGroupAttrItem        `json:"attributes"              binding:"required"`
}
type PatchResCreateGroupData struct {
	Data                      PatchResCreateGroupDataItem    `json:"data"                    binding:"required"`
}
//patch请求，打开apple公链   *************End*************

//请求咱们的itc服务的body，要清理指定Group中的Tester *************Start*************
type ReqDeleteTesterFromClient struct {
	AppId                     string					`json:"app_id"                  binding:"required"`
	GroupId                   string					`json:"group_id"                binding:"required"`
	NumClear				  int						`json:"num_clear"`
	LogSend                   bool                      `json:"log_send"`
	SendUser                  string					`json:"send_user"`
}
//请求咱们的itc服务的body，要清理指定Group中的Tester *************End*************

//苹果TF服务，返回Group中Tester信息 *************Start*************
type ResTesterInfoDataItem struct {
	Type                      string                    `json:"type"                    binding:"required"`
	Id						  string                    `json:"id"                      binding:"required"`
}
type ResTesterInfoData struct {
	Data                      []ResTesterInfoDataItem   `json:"data"                    binding:"required"`
}
//苹果TF服务，返回Group中Tester信息 *************End*************

//上传任意文件到tos上req *************Start*************
type ReqUploadTosFile struct {
	Path  					  string  					`form:"pathName"                binding:"required"`
}
//上传任意文件到tos上req *************End*************

func ReqToAppleTFHasObjMethod(method, url, tokenString string, objReq, objRes interface{}) bool {
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
		if err != nil {
			logs.Info("读取respose的body内容失败")
			return false
		}
		logs.Info("苹果成功返回response\n：%s", string(responseByte))
		if objRes != nil{
			json.Unmarshal(responseByte, objRes)
		}
		return true
	}
}

func ReqToAppleTFHasObjMethodRoutine(method, url, tokenString string, objReq, objRes interface{},ch chan int) bool {
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
		ch <- 2
		return false
	}
	request.Header.Set("Authorization", tokenString)
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		logs.Info("发送请求失败")
		ch <- 2
		return false
	}
	defer response.Body.Close()
	logs.Info("状态码：%d", response.StatusCode)
	if !AssertResStatusCodeOK(response.StatusCode) {
		logs.Info("查看返回状态码")
		logs.Info(string(response.StatusCode))
		responseByte, _ := ioutil.ReadAll(response.Body)
		logs.Info("苹果失败返回response\n：%s", string(responseByte))
		ch <- 2
		return false
	} else {
		responseByte, err := ioutil.ReadAll(response.Body)
		logs.Info("查看苹果的返回值")
		if err != nil {
			logs.Info("读取respose的body内容失败")
			ch <- 2
			return false
		}
		logs.Info("苹果成功返回response\n：%s", string(responseByte))
		if objRes != nil{
			json.Unmarshal(responseByte, objRes)
		}
		ch <- 1
		return true
	}
}
// NO_UPLOAD、（PROCESSING, FAILED, INVALID）、（WAITING_FOR_REVIEW, IN_REVIEW, REJECTED, APPROVED）、（READY_FOR_REVIEW、WAITING_FOR_OTHER_REVIEW、READY_FOR_TEST）
// NO_UPLOAD、PROCESSING、FAILED、INVALID、WAITING_FOR_REVIEW、IN_REVIEW、REJECTED、WAITING_FOR_OTHER_REVIEW 就不要走下一步了，时刻监控此接口

// READY_FOR_REVIEW 走提交审核POST接口，大概率会变成WAITING_FOR_REVIEW状态
// READY_FOR_TEST   走提交审核POST接口，大概率会变成WAITING_FOR_REVIEW状态

// APPROVED可直接调用新建Group接口，新建Group时就填入该版本并打开公链 test
func GetRecentVersionReviewInfo(c *gin.Context)  {
	logs.Info("获取该版本在TF后台的review情况")
	var requestData ReqTFReviewInfoFromClient
	bindJsonError := c.ShouldBindQuery(&requestData)
	if bindJsonError != nil {
		logs.Error("绑定post请求body出错：%v", bindJsonError)
		AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	tokenString := ""
	appAppleId  := ""
	if _, ok := _const.TestFlightAppIdAnd[requestData.AppId]; ok {
		tokenString = GetTokenStringByTeamId(_const.TestFlightAppIdAnd[requestData.AppId]["TeamId"])
		appAppleId  = _const.TestFlightAppIdAnd[requestData.AppId]["AppAppleId"]
	} else {
		logs.Info("客户端发送了未知的app id，不在后段维护的const map中")
		c.JSON(http.StatusNotFound, gin.H{
			"message":   "客户端发送了未知的app id",
			"error_code": "1",
			"data": map[string]interface{}{},
		})
		return
	}
	url := _const.GetBuildAndReviewInfoFromApple + appAppleId
	logs.Info("Token值是：%s",tokenString)
	logs.Info("URL的值是：%s",url)
	var resFromApple ResTFReviewInfoFromApple
	reqResult := ReqToAppleTFHasObjMethod("GET",url,tokenString,nil,&resFromApple)
	if reqResult{
		buildReviewObj := GetReviewStructWithVersion(requestData.VersionLong,requestData.VersionShort,&resFromApple)
		if buildReviewObj.VersionStatus == "READY_FOR_REVIEW" || buildReviewObj.VersionStatus == "READY_FOR_TEST"{
			reviewStatus,respResult := ChangeReadyStatusToReview(buildReviewObj.AppleBuildId,tokenString)
			if respResult{
				buildReviewObj.VersionStatus = reviewStatus
			}else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"message":   "访问苹果提审服务出错",
					"error_code": "3",
					"data": map[string]interface{}{},
				})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "success",
			"error_code": "0",
			"data": buildReviewObj,
		})
		return
	}else {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":   "访问苹果服务出错",
			"error_code": "2",
			"data": map[string]interface{}{},
		})
		return
	}
}

func GetReviewStructWithVersion (longVersion,shortVersion string,resFromApple *ResTFReviewInfoFromApple) *BuildIdOrStatus{
	var versionInfo BuildIdOrStatus
	versionInfo.AppleBuildId = ""
	waitForReview := make([]string,0)
	approved := make([]string,0)
	inReview := make([]string,0)
	rejected := make([]string,0)
	for _,item := range resFromApple.Data{
		if strings.HasPrefix(item.Attributes.LongVersion,shortVersion){
			if longVersion == item.Attributes.LongVersion {
				versionInfo.AppleBuildId = item.BuildId
				if item.Attributes.ProcessingState != "VALID"{
					versionInfo.VersionStatus = item.Attributes.ProcessingState
					return &versionInfo
				}else {
					for _,includeItemCheck := range resFromApple.Included{
						if includeItemCheck.BuildId == item.BuildId {
							versionInfo.VersionStatus = includeItemCheck.ReviewAttributes.BetaReviewState
							return &versionInfo
						}
					}
				}
			}else {
				for _,includeItem := range resFromApple.Included{
					if includeItem.BuildId== item.BuildId {
						if includeItem.ReviewAttributes.BetaReviewState == "WAITING_FOR_REVIEW" {
							waitForReview = append(waitForReview,item.Attributes.LongVersion)
						}else if includeItem.ReviewAttributes.BetaReviewState == "APPROVED"{
							approved = append(approved,item.Attributes.LongVersion)
						}else if includeItem.ReviewAttributes.BetaReviewState == "IN_REVIEW"{
							inReview = append(inReview,item.Attributes.LongVersion)
						}else {
							rejected = append(rejected,item.Attributes.LongVersion)
						}
					}
				}
			}
		}
	}
	if versionInfo.AppleBuildId == "" {
		versionInfo.VersionStatus = "NO_UPLOAD"
	}else {
		if len(waitForReview) == 0 && len(approved) == 0 && len(inReview) == 0{
			versionInfo.VersionStatus = "READY_FOR_REVIEW"
		}else if len(waitForReview) > 0 || len(inReview) > 0 {
			versionInfo.VersionStatus = "WAITING_FOR_OTHER_REVIEW"
		}else if len(approved) >0 {
			versionInfo.VersionStatus = "READY_FOR_TEST"
		}
	}
	return &versionInfo
}

func ChangeReadyStatusToReview (buildId,tokenString string) (string,bool){
	var reqSubm ReqSubmData
	var resSubm ResSubmData
	reqSubm.Data.Type = "betaAppReviewSubmissions"
	reqSubm.Data.Relationships.BuildZone.DataZone.BuildId = buildId
	reqSubm.Data.Relationships.BuildZone.DataZone.BuildType = "builds"
	resResult := ReqToAppleTFHasObjMethod("POST",_const.PutToTFReviewUrl,tokenString,&reqSubm,&resSubm)
	if resResult{
		return resSubm.Data.Attributes.BetaReviewState,resResult
	}else {
		return "appleError",resResult
	}
}

func CreateGroupAddVesion(c *gin.Context) {
	logs.Info("新建TF分组，并且加入版本，打开公链")
	var body ReqTFLinkFromClient
	err := c.ShouldBindJSON(&body)
	utils.RecordError("参数绑定失败", err)
	if err != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	tokenString := ""
	appAppleId := ""
	if _, ok := _const.TestFlightAppIdAnd[body.AppId]; ok {
		tokenString = GetTokenStringByTeamId(_const.TestFlightAppIdAnd[body.AppId]["TeamId"])
		appAppleId = _const.TestFlightAppIdAnd[body.AppId]["AppAppleId"]
	} else {
		logs.Info("客户端发送了未知的app id，不在后段维护的const map中")
		c.JSON(http.StatusNotFound, gin.H{
			"message":    "客户端发送了未知的app id",
			"error_code": "1",
			"data":       map[string]interface{}{},
		})
		return
	}
	var groupCreateReq ReqGroupCreateData
	var groupCreateRes ResCreateGroupData
	groupCreateReq.Data.Type = "betaGroups"
	groupCreateReq.Data.Attributes.GroupName = body.GroupName
	groupCreateReq.Data.Relationships.AppZone.Data.AppleAppId = appAppleId
	groupCreateReq.Data.Relationships.AppZone.Data.AppleAppType = "apps"
	reqResult := ReqToAppleTFHasObjMethod("POST", _const.CreateTFGroupUrl, tokenString, &groupCreateReq, &groupCreateRes)
	if !reqResult {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":    "建立Group分组时，苹果服务出错",
			"error_code": "2",
			"data":       map[string]interface{}{},
		})
		return
	}
	var addBuildReq ReqVersionToGroupData
	var buildItem ReqVersionToGroupItem
	buildItem.Id = body.AppBuildId
	buildItem.Type = "builds"
	addBuildReq.Data = append(addBuildReq.Data,buildItem)
	addBuildToGroupUrl := _const.CreateTFGroupUrl + "/" + groupCreateRes.Data.Id + "/relationships/builds"
	addBuildRes := ReqToAppleTFHasObjMethod("POST",addBuildToGroupUrl,tokenString,&addBuildReq,nil)
	if !addBuildRes{
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":    "把对应版本加入分组时，苹果服务出错",
			"error_code": "3",
			"data":       map[string]interface{}{},
		})
		return
	}
	var patchResLink PatchResCreateGroupData
	var resLink ResCreateGroupData
	patchResLink.Data.Type = "betaGroups"
	patchResLink.Data.Id = groupCreateRes.Data.Id
	patchResLink.Data.Attributes.PublicLinkEnabled = true
	openLinkUrl := _const.CreateTFGroupUrl + "/" + groupCreateRes.Data.Id
	openLinkRes := ReqToAppleTFHasObjMethod("PATCH",openLinkUrl,tokenString,&patchResLink,&resLink)
	if !openLinkRes{
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":    "公链打开时，苹果服务出错",
			"error_code": "4",
			"data":       map[string]interface{}{},
		})
		return
	}else {
		c.JSON(http.StatusOK, gin.H{
			"message":    "success",
			"error_code": "0",
			"data":       &resLink.Data,
		})
		return
	}
}

func TimeOutFunc(timeout chan bool){
	time.Sleep(3 * time.Second)
	timeout <- true
}

//清除指定Group中的Tester
func DeleteGroupTester(c *gin.Context){
	logs.Info("删除指定的TestFlight Group中Tester")
	var body ReqDeleteTesterFromClient
	err := c.ShouldBindJSON(&body)
	utils.RecordError("参数绑定失败", err)
	if err != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败", "failed")
		return
	}
	go func(ReqDeleteTesterFromClient){
		tokenString := ""
		//appAppleId := ""
		successNum := 0
		failNum := 0
		loopNum := 1
		LOOP: for i := 0; true; i++ {
			logs.Info("正在执行第%d次删除任务,请等待",loopNum)
			loopNum ++
			var resTestInfo ResTesterInfoData
			if _, ok := _const.TestFlightAppIdAnd[body.AppId]; ok {
				tokenString = GetTokenStringByTeamId(_const.TestFlightAppIdAnd[body.AppId]["TeamId"])
				//appAppleId = _const.TestFlightAppIdAnd[body.AppId]["AppAppleId"]
			} else {
				textLog := "客户端发送了未知的app id，不在后端维护的const map中"
				logs.Info(textLog)
				SendMessageToMe(&textLog,body.SendUser)
				return
			}
			urlReq := _const.CreateTFGroupUrl + "/" + body.GroupId + "/betaTesters?limit=100"
			//urlReq := _const.TFTesterManagerUrl + "?filter[apps]=" + appAppleId + "&filter[inviteType]=PUBLIC_LINK&limit=100"
			reqResult := ReqToAppleTFHasObjMethod("GET",urlReq,tokenString,nil,&resTestInfo)
			if !reqResult{
				textLog := fmt.Sprintf("访问苹果的人员列表失败，当前成功删除了%d人",successNum)
				logs.Error(textLog)
				SendMessageToMe(&textLog,body.SendUser)
				return
			}
			topLev := ( i + 1 ) * 100
			botLev := i * 100
			brokenSign := false
			if body.NumClear > 0 && (body.NumClear - topLev) < 0 && len(resTestInfo.Data) >= (body.NumClear - botLev) {
				resTestInfo.Data = resTestInfo.Data[0:body.NumClear-botLev]
				brokenSign = true
			}
			timeout := make(chan bool, len(resTestInfo.Data))
			ch := make(chan int,len(resTestInfo.Data))
			for _,item := range resTestInfo.Data {
				go TimeOutFunc(timeout)
				delUrl := _const.TFTesterManagerUrl + "/" + item.Id
				go ReqToAppleTFHasObjMethodRoutine("DELETE",delUrl,tokenString,&resTestInfo,nil,ch)
			}
			for routineNum:=0;routineNum<len(resTestInfo.Data);routineNum++ {
				select {
				case sign := <- ch:
					if sign == 1{
						successNum = successNum + 1
					}else {
						failNum = failNum + 1
					}
				case <- timeout:
					logs.Warn("访问苹果超时")
					failNum = failNum + 1
					break
				}
			}
			if len(resTestInfo.Data) < 100 || brokenSign {
				if failNum > 0 {
					body.NumClear = failNum
					logs.Info("删除收尾时，出现了%d次删除失败的情况",failNum)
					failNum = 0
					logs.Info("即将进行第%d次获取Tester循环，现在等待60秒",loopNum)
					if body.LogSend {
						textLog := fmt.Sprintf("即将进行第%d次获取Tester循环，现在等待60秒，当前删除了%d人", loopNum, successNum)
						SendMessageToMe(&textLog,body.SendUser)
					}
					time.Sleep(61 * time.Second)
					goto LOOP
				}else {
					textLog := fmt.Sprintf("删除TF Tester完毕，当前成功删除了%d人",successNum)
					SendMessageToMe(&textLog,body.SendUser)
					return
				}
			}
			logs.Info("即将进行第%d次获取Tester循环，现在等待60秒",loopNum)
			if body.LogSend {
				textLog := fmt.Sprintf("即将进行第%d次获取Tester循环，现在等待60秒，当前删除了%d人", loopNum, successNum)
				SendMessageToMe(&textLog,body.SendUser)
			}
			time.Sleep(61 * time.Second)
		}
		return
	}(body)
	c.JSON(http.StatusOK, gin.H{
		"message":    "success",
		"error_code": "0",
		"target_delete_num": body.NumClear,
		"data":       map[string]interface{}{},
	})
	return
}

func SendMessageToMe(content *string,userEmail string) {
	botService := service.BotService{}
	botService.SetAppIdAndAppSecret(_const.BotApiId,_const.BotAppSecret)
	openId := "ou_b392d8a6bdcac11e0a7401233cba38ce"
	contentText := form.SendMessageForm{}
	msgType := "text"
	contentText.Content.Text = content
	contentText.MsgType = &msgType
	if userEmail != ""{
		tokenFormservice,err := service.GetTenantAccessToken(_const.BotApiId,_const.BotAppSecret)
		utils.RecordError("获取TenantAccessToken失败", err)
		var reqUserToLark UserInfoReqToLark
		reqUserToLark.Email = userEmail+"@bytedance.com"
		var resUserInfo UserInfoGetFromLark
		reqResult := PostToLarkGetInfo("POST","https://open.feishu.cn/open-apis/user/v4/email2id","Bearer "+tokenFormservice.TenantAccessToken,&reqUserToLark,&resUserInfo)
		if reqResult{
			contentText.OpenID = &resUserInfo.Data.OpenId
			_, errorSendInfo := botService.SendMessage(contentText)
			if errorSendInfo != nil {
				utils.RecordError("bot发送消息错误", errorSendInfo)
			}
		}
	}
	contentText.OpenID = &openId
	_, errorSendInfo := botService.SendMessage(contentText)
	if errorSendInfo != nil {
		utils.RecordError("bot发送消息错误", errorSendInfo)
	}
}

func UploadFileToTos(c *gin.Context)  {
	logs.Info("单独上传文件接口")
	var requestData ReqUploadTosFile
	bindError := c.ShouldBind(&requestData)
	utils.RecordError("绑定post请求body出错：%v", bindError)
	if bindError != nil {
		utils.AssembleJsonResponse(c, http.StatusBadRequest, "请求参数绑定失败，查看是否缺少参数", "failed")
		return
	}
	FileByteInfo, FileFullName := getFileFromRequest(c, "objectFile")
	pathTos := "appleConnectFile/" + requestData.Path + "/" + FileFullName
	logs.Info("路径：%s",pathTos)
	logs.Info("文件内容：%s",string(FileByteInfo))
	var tosBucket = tos.WithAuth(_const.TOS_BUCKET_NAME_JYT, _const.TOS_BUCKET_TOKEN_JYT)
	context, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	tosPutClient, err := tos.NewTos(tosBucket)
	err = tosPutClient.PutObject(context, pathTos, int64(len(FileByteInfo)), bytes.NewBuffer(FileByteInfo))
	if err != nil {
		logs.Error("tos上传错误的收录:%s",err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"message":    "error",
			"error_code": "1",
		})
		return
	}else {
		c.JSON(http.StatusOK, gin.H{
			"message":    "success",
			"error_code": "0",
		})
		return
	}

}
