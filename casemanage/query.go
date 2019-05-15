package casemanage

import (
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/const"
	"code.byted.org/gopkg/logs"
	"code.byted.org/gopkg/tos"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strconv"
	"path/filepath"
	"code.byted.org/clientQA/itc-server/detect"
	"os"
	"io"
	"time"
	"fmt"
	"context"
	"math/rand"
	"bytes"
	"encoding/json"
	"strings"
)

/*
更新数据输入结构
 */
type EditStruct struct {
	Id 				int  		`json:"id"`
	Solution 		string		`json:"solution"`
}
/*
删除数据输入结构
 */
type DeleteStruct struct {
	Id 				int  		`json:"id"`
}


/*
	query rejCases with conditions
 */


func GetRejCasesByConditions(c *gin.Context){
	pageS,ok := c.GetQuery("page")
	if !ok{
		logs.Error("缺少page参数！")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少page参数！",
			"errorCode" : -1,
			"total":0,
			"data" : "缺少page参数！",
		})
		return
	}
	page,err := strconv.Atoi(pageS)
	if (err !=nil || page<=0){
		logs.Error("page参数不符合要求")
		c.JSON(http.StatusOK, gin.H{
			"message" : "page参数不符合要求！",
			"errorCode" : -1,
			"total":0,
			"data" : "page参数不符合要求！",
		})
		return
	}

	pageSizeS,ok2 := c.GetQuery("pageSize")
	if !ok2 {
		logs.Error("缺少pageSize参数")
		c.JSON(http.StatusOK,gin.H{
			"message":"缺少pageSize参数",
			"errorCode" : -1,
			"total":0,
			"data" : "缺少pageSize参数！",
		})
		return
	}
	pageSize,err2 := strconv.Atoi(pageSizeS)
	if (err2 != nil|| pageSize <= 0){
		logs.Error("pageSize参数不符合要求")
		c.JSON(http.StatusOK, gin.H{
			"message" : "pageSize参数不符合要求！",
			"errorCode" : -1,
			"total":0,
			"data" : "pageSize参数不符合要求！",
		})
		return
	}
	var param map[string]string
	param = make(map[string]string)
	appId, ok := c.GetQuery("appId")
	version,ok2 :=c.GetQuery("version")


	//query with conditions
	condition:=""
	if ok{
		appIdInt,err := strconv.Atoi(appId)
		if err != nil {
			logs.Error("appId参数不符合要求")
			c.JSON(http.StatusOK, gin.H{
				"message" : "appId参数不符合要求！",
				"errorCode" : -1,
				"total":0,
				"data" : "appId参数不符合要求！",
			})
			return
		}
		condition+=" app_id="+fmt.Sprint(appIdInt)
	}
	if ok2{
		if ok {
			condition+=" and version like '%"+version+"%'"
		}else{
			condition+=" version like '%"+version+"%'"
		}

	}

	param["condition"] = condition
	param["page"] = strconv.Itoa(page)
	param["pageSize"] = strconv.Itoa(pageSize)
	logs.Info("before param:%v",param)

	items,total,err := dal.QueryByConditions(param)
	if err!=nil{
		c.JSON(http.StatusOK,gin.H{
			"message":"数据库操作失败",
			"errorCode":-1,
			"total":total,
			"data":err,
		})
		return
	}
	logs.Info("query with condition success")
	c.JSON(http.StatusOK, gin.H{
		"message":"success",
		"errorCode":0,
		"total":total,
		"data":*items,
	})
	return

}


/*
	add a new rejCase
 */
func AddRejCase(c *gin.Context)  {
	form,err := c.MultipartForm()
	if err != nil {
		logs.Error("info get failed! %v",err)
		c.JSON(http.StatusOK,gin.H{
			"errorCode":-1,
			"message":"获取post信息失败！",
		})
		return
	}

	var r dal.RejInfo
		r.AppId,err = strconv.Atoi(form.Value["appId"][0])
		r.AppName = form.Value["appName"][0]
		r.RejRea = form.Value["rejRea"][0]
		r.RejTime,_ = time.ParseInLocation("2006-01-02",form.Value["rejTime"][0],time.Local)
		r.Solution = form.Value["solution"][0]
		r.Version = form.Value["version"][0]

	appId := r.AppId

	files := form.File["uploadFile"]
	_tmpDir := "./tmp/"+strconv.Itoa(appId)
	exist, err := detect.PathExists(_tmpDir)
	if !exist{
		os.MkdirAll(_tmpDir, os.ModePerm)
	}
	var path = ""
	if(len(files)>0){
		for _,file := range files{
			var filename = file.Filename
			fileNameInfo := strings.Split(filename,".")

			fileReal,err := file.Open()
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"message":"图片处理失败，请联系相关人员！",
					"errorCode":-1,
				})
				logs.Fatal("临时图片文件创建失败")
				return
			}
			defer fileReal.Close()
			filepath := _tmpDir + "/"+fmt.Sprint(time.Now().UnixNano())+"."+fileNameInfo[len(fileNameInfo)-1]
			out,err := os.Create(filepath)
			defer out.Close()
			if err != nil{
				c.JSON(http.StatusOK, gin.H{
					"message":"图片处理失败，请联系相关人员！",
					"errorCode":-1,
				})
				logs.Fatal("临时图片文件创建失败%v",err)
				return
			}
			_, err = io.Copy(out, fileReal)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"message":"安装包文件处理失败，请联系相关人员！",
					"errorCode":-1,
				})
				logs.Fatal("临时图片保存失败%v",err)
				return
			}

			returnUrl,err := Upload2Tos(filepath)
			if err != nil {
				logs.Error("图片上传tos失败%v",err)
				c.JSON(http.StatusOK, gin.H{
					"message":"图片上传tos失败",
					"errorCode":-1,
				})
				return
			}

			path+=filename+"--"+returnUrl+";"
			os.Remove(filepath)
		}
	}else{
		logs.Info("-------------no files")
	}
	var data = make(map[string]interface{})
	data["info"] = r
	data["picPath"] = path

	errs := dal.InsertRejCase(data)
	if errs != nil {
		c.JSON(http.StatusOK,gin.H{
			"errorCode":-1,
			"message":errs,
		})
		return
	}else {
		logs.Info("add success")
		c.JSON(http.StatusOK,gin.H{
			"errorCode":0,
			"message":"success",
		})
	}
	return

}

/*
	delete a rejCase
 */
func DeleteRejCase(c *gin.Context)  {
	param,_ := ioutil.ReadAll(c.Request.Body)
	var info DeleteStruct
	err := json.Unmarshal(param,&info)
	if err != nil {
		logs.Error("Id 格式不正确，%v",err)
		c.JSON(http.StatusOK, gin.H{
			"message" : "id格式不正确！",
			"errorCode" : -1,
		})
		return
	}

	result := dal.DeleteCase(info.Id)
	if result != nil {
		c.JSON(http.StatusOK,gin.H{
			"message":result,
			"errorCode":-1,
		})
		return
	}else{
		logs.Info("delete success")
		c.JSON(http.StatusOK,gin.H{
			"message":"success",
			"errorCode":0,
		})
	}
	return

}

func EditRejCaseofSolution(c *gin.Context)  {
	param, _ := ioutil.ReadAll(c.Request.Body)
	var r EditStruct
	err := json.Unmarshal(param, &r)
	if err != nil {
		logs.Error("参数格式错误,%v ", err)
		c.JSON(http.StatusOK, gin.H{
			"message" : "提交参数格式错误",
			"errorCode" : -1,
		})
		return
	}

	var data = make(map[string]string)
	data["condition"] = "id="+strconv.Itoa(r.Id)
	data["solution"] = r.Solution
	result := dal.UpdateRejCaseofSolution(data)
	if result != nil {
		c.JSON(http.StatusOK,gin.H{
			"message":result,
			"errorCode":-1,
		})
		return
	}
	logs.Info("edit success")
	c.JSON(http.StatusOK,gin.H{
		"errorCode":0,
		"message":"success",
	})
	return
}


func Upload2Tos(path string) (string, error){

	var tosBucket = tos.WithAuth(_const.TOS_BUCKET_NAME, _const.TOS_BUCKET_KEY)
	context, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	tosPutClient, err := tos.NewTos(tosBucket)
	fileName := filepath.Base(path)
	byte, err := ioutil.ReadFile(path)
	if err != nil {
		logs.Error("%s", "打开文件失败" + err.Error())
		return "",err
	}
	key := fileName
	logs.Info("key: " + key)
	err = tosPutClient.PutObject(context, key, int64(len(byte)), bytes.NewBuffer(byte))
	if err != nil {
		logs.Error("%s", "上传tos失败：" + err.Error())
		return "",err
	}
	domains := tos.GetDomainsForLargeFile("TT", path)
	domain := domains[rand.Intn(len(domains)-1)]
	domain = "tosv.byted.org/obj/" + _const.TOS_BUCKET_NAME
	var returnUrl string
	returnUrl = "https://" + domain + "/" + key

	return returnUrl, nil
}