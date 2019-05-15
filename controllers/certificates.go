package controllers

import (
	"bytes"
	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/tos"
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

/*
遍历证书，查询过期日期并提醒给拥有者
*/
func CertificateController(c *gin.Context){
	if certificateSet := dal.QueryCertificate(map[string]interface{}{}); len(*certificateSet) != 0 {

		var expiredCertificates  []map[string]interface{}  //存放即将过期证书记录
		copTime := time.Now().Unix()
		//遍历数据库，找到符合条件的记录
		for _, certificate := range *certificateSet{
			expireTime := certificate.ExpireTime
			copExpire, _ := strconv.ParseInt(expireTime,10,64)
			subTime :=  copExpire - copTime
			if subTime <= 2592000000 && subTime > 0{
				appName := certificate.Appname
				usage := certificate.Usage
				creator := certificate.Creator
				//发送过期信息的用户列表
				var newMails []string
				newMails = append(newMails, strings.Replace(creator, " ", "", -1))
				newMails = append(newMails, "gongrui")
				newMails = append(newMails, "chenyujun")
				itemMap := map[string]interface{}{
					"appname":appName,
					"usage":usage,
					"larkPeople":newMails,
					"expire_time":expireTime,
				}
				expiredCertificates = append(expiredCertificates, itemMap)
			}
		}
		//符合条件的证书给creator发lark提醒
		for _, item := range expiredCertificates{
			appname := item["appname"]
			usage := item["usage"]
			//unix时间戳转换为北京时间，提示给用户
			expireTime, _ := strconv.ParseInt(item["expire_time"].(string), 10, 64)
			noticeExpireTime := time.Unix(expireTime, 0).Format("02/01/2006 15:04:05")

			larkPeople := item["larkPeople"]
			for _, people := range larkPeople.([]string){
				larkMessage := "预审平台系统消息:\n应用:" + appname.(string) + "\n用途：" +  usage.(string) + "\n过期时间" + noticeExpireTime
				utils.LarkDingOneInner(people, larkMessage)
				fmt.Println("people: %s, larkMessage: %s", people, larkMessage)
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"message" : "success",
		"errorCode" : 0,
		"data" : "即将过期证书已经全部提醒给创建者啦～～～",
	})
}

/*
数据库查询符合条件的证书记录，并分页返回数据
前端传入参数：
appName：
type
creator
page
pageSize
appid
*/
func GetCertificates(c *gin.Context) {
	//获取用户信息
	name, f := c.Get("username")
	if !f {
		c.JSON(http.StatusOK, gin.H{
			"message" : "未获取到用户信息！",
			"errorCode" : -1,
			"data" : "未获取到用户信息！",
		})
		return
	}
	var certificate []dal.CertificateModel
	connection, err := database.GetConneection()
	if err != nil {
		logs.Error("Connect to Db failed: %v", err)
		return
	}
	defer connection.Close()

	db := connection.Table(dal.CertificateModel{}.TableName()).LogMode(_const.DB_LOG_MODE)
	if appName, isExit := c.GetQuery("appName"); isExit == true{
		//中文添加特殊处理，暂时还是无法查询中文记录
		appname := "%"
		appname += string(appName[:])
		appname += "%"
		fmt.Println(reflect.TypeOf(appname))
		db = db.Where("appname LIKE ?", appname)
	}
	if cerType, isExit := c.GetQuery("type"); isExit == true{
		db = db.Where("type = ?", cerType)
	}
	if isSelectCreator, isExit := c.GetQuery("creator"); isExit == true && isSelectCreator == "on"{
		db = db.Where("creator = ?", name)
	}

	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize"))

	if page > 0 && pageSize > 0 {
		db = db.Limit(pageSize).Offset((page - 1) * pageSize)
	}

	if err := db.Find(&certificate).GetErrors(); err != nil {
		for _, singleError := range err{
			logs.Error(singleError.Error())
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": certificate,
	})
}

/*
添加证书
*/
func AddCertificate(c *gin.Context){
	name, f := c.Get("username")
	if !f {
		c.JSON(http.StatusOK, gin.H{
			"message" : "未获取到用户信息！",
			"errorCode" : -1,
			"data" : "未获取到用户信息！",
		})
		return
	}

	file, header, _ := c.Request.FormFile("certificate_file")
	if file == nil {
		c.JSON(http.StatusOK, gin.H{
			"message":"未选择上传的文件！",
		})
		logs.Error("未选择上传的文件！")
		return
	}
	defer file.Close()

<<<<<<< HEAD
	certificateFileName := header.Filename //获取证书名称
=======
	certificateFileName := header.Filename  //获取证书名称
>>>>>>> ITC证书管理功能迁移
	pem := c.PostForm("pem")        //是否需要返回pem_file标志

	//查询证书过期日期
	response := func() (resp *http.Response){

		upstreamUrl := "http://10.2.9.226:9527/query_certificate_expire_date"

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("certificate", certificateFileName)
		if err != nil {
			logs.Error("访问过期日期POST请求create form file错误！", err.Error())
		}
		//这里的io.Copy实现,会把file文件都读取到内存里面，然后当做一个buffer传给NewRequest. 对于大文件来说会占用很多内存
		if _, err = io.Copy(part, file); err != nil{
			logs.Error("访问过期日期POST请求复制文件错误！", err.Error())
		}

		if c.PostForm("type") == ".p12"{
			var export_pem string
			if pem == "on"{
				export_pem = "1"
			}else{
				export_pem = "0"
			}
			writer.WriteField("export_pem", export_pem)
			writer.WriteField("pass", c.PostForm("password"))
		}
		writer.WriteField("username", name.(string))
		writer.WriteField("type", c.PostForm("type"))
		contentType := writer.FormDataContentType()
		if err = writer.Close(); err!= nil{
			logs.Error("关闭writer出错！", err.Error())
		}
		if response, err := http.Post(upstreamUrl, contentType, body); err != nil{
			logs.Error("获取过期信息失败！", err.Error())
		}else{
			return response
		}
		return
	}()

	//得到证书日期，pem等信息后存入数据库
<<<<<<< HEAD
=======

>>>>>>> ITC证书管理功能迁移
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil{
		logs.Error("访问处理失败！", err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "证书处理失败！可能是证书本身的格式问题，或者是密码错误",
		})
	}

	result := make(map[string]interface{})
	json.Unmarshal(responseBody, &result)

	if len(result) == 1{
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "证书处理失败！可能是证书本身的格式问题，或者是密码错误",
		})
		return
	}

	//db Model
	var certificateModel dal.CertificateModel
	certificateModel.Creator = name.(string)
	certificateModel.ExpireTime = strconv.FormatFloat(result["expire_time"].(float64), 'E', -1, 64)
	certificateModel.Appname = c.PostForm("appName")
	certificateModel.AppId, _ = strconv.Atoi(c.PostForm("appId"))
	certificateModel.Usage = c.PostForm("application")
	certificateModel.Mails = c.PostForm("mails")
	certificateModel.Type = c.PostForm("type")
	certificateModel.Password = c.PostForm("password")
	certificateModel.CertificateFileName = certificateFileName

	//如果pem=on，返回中包含pem_file字符串
	if _, ok := result["pem_file"]; ok{
		pem_file_name := strings.Split(certificateFileName, ".")[0] +".pem"

		certificateModel.PemFile = result["pem_file"].(string)
		certificateModel.PemFileName = pem_file_name
	}

	//证书上传到TOS，获取访问地址并存到数据库certificate_file那一列
	localCertificatePath := "./"+ c.PostForm("appId") + "." + strings.Split(certificateFileName, ".")[1]
	newFile, err := os.Create(localCertificatePath)   //本地创建临时文件
	if err!= nil{
		logs.Error("创建临时失败！", err.Error())
	}
	certificateFile, _ := header.Open()
	io.Copy(newFile, certificateFile)  //复制证书文件内容到临时文件

	//上传证书到TOS，并获取访问地址
	fileTosUrl := uploadTos(localCertificatePath)
	certificateModel.CertificateFile = fileTosUrl

	if err = os.Remove(localCertificatePath); err != nil{  //删除本地创建的临时文件
		logs.Error("删除临时文件失败！", err.Error())
	}

	if dal.InsertCertificate(certificateModel){
		c.JSON(http.StatusOK, gin.H{
			"message": "OK!",
		})
	}else{
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Failed!",
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
		logs.Error("%s", "打开文件失败" + err.Error())

	}
	key := fmt.Sprint(time.Now().UnixNano()) + "_" + fileName
	logs.Info("key: " + key)
	err = tosPutClient.PutObject(context, key, int64(len(byte)), bytes.NewBuffer(byte))
	if err != nil {
		logs.Error("%s", "上传tos失败：" + err.Error())
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


