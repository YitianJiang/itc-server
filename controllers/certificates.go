package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"code.byted.org/clientQA/itc-server/detect"

	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/tos"
	"github.com/gin-gonic/gin"
)

/*
遍历证书，查询过期日期并提醒给拥有者
*/
func CertificateController(c *gin.Context) {
	if certificateSet := dal.QueryCertificate(map[string]interface{}{}); len(*certificateSet) != 0 {
		var expiredCertificates []map[string]interface{} //存放即将过期证书记录
		copTime := time.Now().Unix()
		//遍历数据库，找到符合条件的记录
		for _, certificate := range *certificateSet {
			expireTime := certificate.ExpireTime
			copExpire, _ := strconv.ParseInt(expireTime, 10, 64)
			subTime := copExpire - copTime
			if subTime <= 2592000000 && subTime > 0 {
				appName := certificate.Appname
				usage := certificate.Usage
				creator := certificate.Creator
				//发送过期信息的用户列表
				var newMails []string
				newMails = append(newMails, strings.Replace(creator, " ", "", -1))
				newMails = append(newMails, "gongrui")
				newMails = append(newMails, "chenyujun")
				newMails = append(newMails, "kanghuaisong")
				newMails = append(newMails, "zhangshuai.02")
				itemMap := map[string]interface{}{
					"appname":     appName,
					"usage":       usage,
					"larkPeople":  newMails,
					"expire_time": expireTime,
				}
				expiredCertificates = append(expiredCertificates, itemMap)
			}
		}
		//符合条件的证书给creator发lark提醒
		for _, item := range expiredCertificates {
			appname := item["appname"]
			usage := item["usage"]
			//unix时间戳转换为北京时间，提示给用户
			expireTime, _ := strconv.ParseInt(item["expire_time"].(string), 10, 64)
			noticeExpireTime := time.Unix(expireTime, 0).Format("02/01/2006 15:04:05")
			larkPeople := item["larkPeople"]
			for _, people := range larkPeople.([]string) {
				larkMessage := "预审平台系统消息:\n应用:" + appname.(string) + "\n用途：" + usage.(string) + "\n过期时间" + noticeExpireTime
				utils.LarkDingOneInner(people, larkMessage)
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message":   "success",
		"errorCode": 0,
		"data":      "即将过期证书已经全部提醒给创建者啦～～～",
	})
}

/*
数据库查询符合条件的证书记录，并分页返回数据
前端传入参数：
appName：app名称
type：   证书类型
creator  创建者
page     第几页
pageSize 页面展示数量
appid    APP ID
*/
func GetCertificates(c *gin.Context) {
	//获取用户信息
	name, f := c.Get("username")
	if !f {
		c.JSON(http.StatusOK, gin.H{
			"message":   "未获取到用户信息！",
			"errorCode": -1,
			"data":      "未获取到用户信息！",
		})
		return
	}
	queryMap := make(map[string]interface{})
	//参数必填检验
	if appName, isExit := c.GetQuery("appName"); isExit == true {
		queryMap["appname"] = appName
	}
	if appId, isExit := c.GetQuery("appId"); isExit == true {
		queryMap["appId"] = appId
	}
	if cerType, isExit := c.GetQuery("type"); isExit == true {
		queryMap["certificate_style"] = cerType
	}
	if isSelectCreator, isExit := c.GetQuery("creator"); isExit == true && isSelectCreator == "on" {
		queryMap["creator"] = name //查询"我"创建的证书信息
	}
	if user, isExit := c.GetQuery("user"); isExit == true {
		if _, ok := queryMap["creator"]; ok {
			c.JSON(http.StatusOK, gin.H{
				"message":   "creator和user不能同时查询!",
				"errorCode": -1,
				"total":     0,
				"data":      []map[string]interface{}{},
			})
			return
		}
		queryMap["creator"] = user //查询user用户创建的证书信息
	}
	totalCertificates := dal.QueryCertificate(queryMap)
	//存储符合条件的数据库总条数
	var total int
	if totalCertificates != nil {
		total = len(*totalCertificates)
	}
	if page, isExit := c.GetQuery("page"); isExit == true {
		queryMap["page"] = page
	}
	if pageSize, isExit := c.GetQuery("pageSize"); isExit == true {
		queryMap["pageSize"] = pageSize
	}
	certificate := dal.QueryLikeCertificate(queryMap)
	//查询符合条件的数据,Json转换返回标准形式
	var downloadWhitePeople = map[string]int{ //下载白名单
		"zhangshuai.02":1,
		"gongrui":1,
		"kanghuaisong":1,
		"yinzhihong":1}
	var data []map[string]interface{}
	for _, cer := range *certificate {
		certificateTemp, err1 := json.Marshal(cer)
		certificateRes := make(map[string]interface{})
		err2 := json.Unmarshal(certificateTemp, &certificateRes)
		if _, ok:= downloadWhitePeople[name.(string)]; name.(string) != certificateRes["creator"] && !ok{
			certificateRes["certificateFile"] = "***"//不在白名单中隐藏下载url
		}
		data = append(data, certificateRes)
		if err1 != nil || err2 != nil {
			logs.Error("数据库结果转成json转成map出错！", err1.Error(), err2.Error())
		}
	}
	if certificate != nil && len(*certificate) > 0 {
		c.JSON(http.StatusOK, gin.H{
			"message":   "OK!",
			"errorCode": 0,
			"total":     total,
			"data":      data,
		})
	} else {
		//返回空
		c.JSON(http.StatusOK, gin.H{
			"message":   "未查询到符合条件的数据!",
			"errorCode": 0,
			"total":     total,
			"data":      []map[string]interface{}{},
		})
	}
}

/*
添加证书
*/
func AddCertificate(c *gin.Context) {
	name, f := c.Get("username")
	if !f {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -1,
			"message":   "未获取到用户信息！",
		})
		return
	}
	//获取上传的文件
	file, header, _ := c.Request.FormFile("certificateFile")
	if file == nil {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -1,
			"message":   "未选择上传的文件！",
		})
		logs.Error("未选择上传的文件！")
		return
	}
	defer file.Close()
	certificateFileName := header.Filename //获取证书名称
	//获取参数
	pem, isExitPem := c.GetPostForm("pem")
	pass, isExitPass := c.GetPostForm("password")
	fileType, isExitType := c.GetPostForm("type")
	appName, isExitAppName := c.GetPostForm("appName")
	appId, isExitAppId := c.GetPostForm("appId")
	usage, isExitUsage := c.GetPostForm("usage")
	mails, isExitMails := c.GetPostForm("mails")
	//判空处理,只有.p12证书需要password和pem字段
	if isExitType == false || (isExitType == true && (fileType == ".p12" && (isExitPem == false || isExitPass == false))) || isExitAppName == false || isExitUsage == false || isExitMails == false || isExitAppId == false {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -1,
			"message":   "上传文件参数填写不完整！",
		})
		logs.Error("上传文件参数填写不完整！")
		return
	}
	//查询证书过期日期
	response := func() (resp *http.Response) {
		upstreamUrl := "http://"+detect.DETECT_URL_PRO+"/query_certificate_expire_date" //过期日期访问地址
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("certificate", certificateFileName)
		if err != nil {
			logs.Error("访问过期日期POST请求create form file错误！", err.Error())
		}
		//这里的io.Copy实现,会把file文件都读取到内存里面，然后当做一个buffer传给NewRequest. 对于大文件来说会占用很多内存
		if _, err = io.Copy(part, file); err != nil {
			logs.Error("访问过期日期POST请求复制文件错误！", err.Error())
		}
		if fileType == ".p12" {
			var export_pem string
			if pem == "on" {
				export_pem = "1"
			} else {
				export_pem = "0"
			}
			writer.WriteField("export_pem", export_pem)
			writer.WriteField("pass", pass)
		}
		writer.WriteField("username", name.(string))
		writer.WriteField("type", fileType)
		contentType := writer.FormDataContentType()
		if err = writer.Close(); err != nil {
			logs.Error("关闭writer出错！", err.Error())
		}
		response, err := http.Post(upstreamUrl, contentType, body)
		if err != nil {
			logs.Error("获取证书过期信息失败！", err.Error())
		}
		return response
	}()
	//返回response判空处理
	if response == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode": -1,
			"message":   "访问过期日期返回为空",
		})
		return
	}
	//得到证书日期，pem等信息后存入数据库
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.Error("返回Response处理失败！", err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode": -1,
			"message":   "返回Response处理失败！",
		})
		return
	}
	result := make(map[string]interface{})
	json.Unmarshal(responseBody, &result)
	if len(result) == 1 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"errorCode": -1,
			"message":   "返回结果错误，没有证书过期日期信息",
		})
		return
	}
	//db Model
	var certificateModel dal.CertificateModel
	certificateModel.Creator = name.(string)
	certificateModel.ExpireTime = strconv.FormatInt(int64(result["expire_time"].(float64)), 10)
	certificateModel.Appname = appName
	certificateModel.AppId, _ = strconv.Atoi(appId)
	certificateModel.Usage = usage
	certificateModel.Mails = mails
	certificateModel.Type = fileType
	certificateModel.CertificateFileName = certificateFileName
	if fileType == ".p12" {
		certificateModel.Password = pass
		//如果pem=on，返回中包含pem_file字符串
		if _, ok := result["pem_file"]; ok {
			pem_file_name := strings.Split(certificateFileName, ".")[0] + ".pem"
			certificateModel.PemFileName = pem_file_name
			certificateModel.PemFile = result["pem_file"].(string)
		}
	}
	//证书上传到TOS，获取访问地址并存到数据库certificate_file那一列
	localCertificatePath := "./" + c.PostForm("appId") + "." + strings.Split(certificateFileName, ".")[1]
	newFile, err := os.Create(localCertificatePath) //本地创建临时文件
	if err != nil {
		c.JSON(http.StatusCreated, gin.H{
			"errorCode": -1,
			"message":   "创建临时文件失败！",
		})
		logs.Error("创建临时失败！", err.Error())
		return
	}
	certificateFile, _ := header.Open()
	io.Copy(newFile, certificateFile) //复制证书文件内容到临时文件
	//上传证书到TOS，并获取访问地址
	fileTosUrl := uploadTos(localCertificatePath)
	if fileTosUrl == "" {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": -1,
			"message":   "得到的tos存储地址为空!",
		})
		logs.Error("得到的tos存储地址为空!")
		return
	}
	certificateModel.CertificateFile = fileTosUrl
	if err = os.Remove(localCertificatePath); err != nil { //删除本地创建的临时文件
		logs.Error("删除临时文件失败！", err.Error())
	}
	if dal.InsertCertificate(certificateModel) {
		c.JSON(http.StatusOK, gin.H{
			"errorCode": 0,
			"message":   "OK!",
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"errorCode": -1,
			"message":   "数据库存储错误，新建证书失败！",
		})
	}
}

/*
文件上传到TOS，并获取访问url
*/
func uploadTos(path string) string {
	var tosBucket = tos.WithAuth(_const.TOS_BUCKET_NAME, _const.TOS_BUCKET_KEY)
	context, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	tosPutClient, err := tos.NewTos(tosBucket)
	fileName := filepath.Base(path)
	byte, err := ioutil.ReadFile(path)
	if err != nil {
		logs.Error("%s", "打开文件失败"+err.Error())
	}
	key := strconv.FormatInt(time.Now().UnixNano(), 10) + "_" + fileName
	logs.Info("key: " + key)
	err = tosPutClient.PutObject(context, key, int64(len(byte)), bytes.NewBuffer(byte))
	if err != nil {
		logs.Error("%s", "上传tos失败："+err.Error())
		return ""
	}
	domains := tos.GetDomainsForLargeFile("TT", path)
	domain := domains[rand.Intn(len(domains)-1)]
	domain = "tosv.byted.org/obj/" + _const.TOS_BUCKET_NAME
	var returnUrl string
	returnUrl = "https://" + domain + "/" + key
	logs.Info("returnUrl: " + returnUrl)
	return returnUrl
}
