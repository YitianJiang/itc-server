package casemanage

import (
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/gopkg/logs"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strconv"
	"path/filepath"
)


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
	if err !=nil {
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
	if err2 != nil{
		logs.Error("pageSize参数不符合要求")
		c.JSON(http.StatusOK, gin.H{
			"message" : "pageSize参数不符合要求！",
			"errorCode" : -1,
			"total":0,
			"data" : "pageSize参数不符合要求！",
		})
		return
	}
	var param map[string]interface{}
	param = make(map[string]interface{})
	appId, ok := c.GetQuery("appId")
	version,ok2 :=c.GetQuery("version")

	//no conditions,query all rejCases
	//if !(ok||ok2){
	//	logs.Info("no condition,query all rejCases!")
	//	items,total,err:=dal.QueryAllRejCases(page,pageSize)
	//	if err!=nil{
	//		c.JSON(http.StatusOK,gin.H{
	//			"message":"数据库操作失败",
	//			"errorCode":-1,
	//			"total":total,
	//			"data":err,
	//		})
	//		return
	//	}
	//
	//	c.JSON(http.StatusOK, gin.H{
	//		"message":"success",
	//		"errorCode":0,
	//		"total":total,
	//		"data":*items,
	//	})
	//	return
	//}

	//query with conditions
	condition:=""
	if ok{
		condition+=" appId="+appId
	}
	if ok2{
		condition+=" version="+version
	}
	param["condition"] = condition
	param["page"] = page
	param["pageSize"] = pageSize

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
get the list of rejCases (all)
 */
//func getAllRejCases(c *gin.Context){
//
//	pageS,ok := c.GetQuery("page")
//	if !ok{
//		logs.Error("缺少page参数！")
//		c.JSON(http.StatusOK, gin.H{
//			"message" : "缺少page参数！",
//			"errorCode" : -1,
//			"total":0,
//			"data" : "缺少page参数！",
//		})
//		return
//	}
//	page,err := strconv.Atoi(pageS)
//	if err !=nil {
//		logs.Error("page参数不符合要求")
//		c.JSON(http.StatusOK, gin.H{
//			"message" : "page参数不符合要求！",
//			"errorCode" : -1,
//			"total":0,
//			"data" : "page参数不符合要求！",
//		})
//		return
//	}
//
//	pageSizeS,ok2 := c.GetQuery("pageSize")
//	if !ok2 {
//		logs.Error("缺少pageSize参数")
//		c.JSON(http.StatusOK,gin.H{
//			"message":"缺少pageSize参数",
//			"errorCode" : -1,
//			"total":0,
//			"data" : "缺少pageSize参数！",
//		})
//		return
//	}
//	pageSize,err2 := strconv.Atoi(pageSizeS)
//	if err2 != nil{
//		logs.Error("pageSize参数不符合要求")
//		c.JSON(http.StatusOK, gin.H{
//			"message" : "pageSize参数不符合要求！",
//			"errorCode" : -1,
//			"total":0,
//			"data" : "pageSize参数不符合要求！",
//		})
//		return
//	}
//
//	items,total,err:=dal.QueryAllRejCases(page,pageSize)
//	if err!=nil{
//		c.JSON(http.StatusOK,gin.H{
//			"message":"数据库操作失败",
//			"errorCode":-1,
//			"total":total,
//			"data":err,
//		})
//		return
//	}
//	logs.Info("query success")
//	c.JSON(http.StatusOK, gin.H{
//		"message":"success",
//		"errorCode":0,
//		"total":total,
//		"data":*items,
//	})
//	return
//}

/*
	add a new rejCase
 */
func AddRejCase(c *gin.Context)  {
	err := c.Request.ParseMultipartForm(1 << 20)
	if err != nil {
		logs.Error("info get failed!")
		c.JSON(http.StatusOK,gin.H{
			"errorCode":-1,
			"message":"获取post信息失败！",
		})
		return
	}
	var r dal.RejInfo
	if c.Request.MultipartForm != nil {
		r.AppId = c.Request.MultipartForm.Value["appId"]
		r.AppName = c.Request.MultipartForm.Value["appName"]
		r.RejRea = c.Request.MultipartForm.Value["rejRea"]
		r.RejTime = c.Request.MultipartForm.Value["rejTime"]
		r.Solution = c.Request.MultipartForm.Value["solution"]
	}
	//param, _ := ioutil.ReadAll(c.Request.Body)
	//var r dal.RejInfo
	//err := json.Unmarshal(param, &r)
	//if err != nil {
	//	logs.Error("参数格式错误,%v ", err)
	//	c.JSON(http.StatusOK, gin.H{
	//		"message" : "提交参数格式错误",
	//		"errorCode" : -1,
	//	})
	//	return
	//}
	//
	//err := c.Request.ParseMultipartForm(1 << 20)
	//if err != nil {
	//	logs.Error("photo upload failed!")
	//	c.JSON(http.StatusOK,gin.H{
	//		"errorCode":-1,
	//		"message":"图片上传失败！",
	//		"data":err,
	//	})
	//	return
	//}
	appId := r.AppId
	//if appId == 0 {
	//	logs.Error("appId参数不合法")
	//	c.JSON(http.StatusOK, gin.H{
	//		"message" : "appId参数不合法",
	//		"errorCode" : -1,
	//	})
	//	return
	//}
	files := c.Request.MultipartForm.File["uploadFile"]
	_tmpDir := "./tmp/"+appId
	exist, err := PathExists(_tmpDir)
	if !exist{
		os.Mkdir(_tmpDir, os.ModePerm)
	}

	var path = ""
	if(len(files)>0){
		for _,file := range files{
			var filename = file.Filename
			filepath := _tmpDir + "/"+filename
			out,err := os.Create(filepath)
			defer out.Close()
			if err != nil{
				c.JSON(http.StatusOK, gin.H{
					"message":"图片处理失败，请联系相关人员！",
					"errorCode":-1,
				})
				logs.Fatal("临时图片文件创建失败")
				return
			}
			_, err = io.Copy(out, file)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{
					"message":"安装包文件处理失败，请联系相关人员！",
					"errorCode":-1,
				})
				logs.Fatal("临时图片保存失败")
				return
			}

			returnUrl,err := upload2Tos(filepath)
			if err != nil {
				logs.Error("图片上传tos失败")
				c.JSON(http.StatusOK, gin.H{
					"message":"图片上传tos失败",
					"errorCode":-1,
				})
				return
			}
			//outputPath := "./rejCase/"+strconv.Itoa(appId)+"/"+file.Filename
			//if err := c.SaveUploadedFile(file, outputPath); err != nil {
			//	logs.Error("pic %s upload to tos failed!",file.Filename)
			//	c.JSON(http.StatusOK,gin.H{
			//		"errorCode":-1,
			//		"message":"图片上传服务器失败！",
			//		"data":err,
			//	})
			//	return
			//}
			path+=filename+"--"+returnUrl+";"
			os.Remove(filepath)
		}
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
	idS,ok := c.GetQuery("id")
	if !ok {
		logs.Error("no ID")
		c.JSON(http.StatusOK, gin.H{
			"message" : "缺少ID参数！",
			"errorCode" : -1,
		})
		return
	}
	id,err := strconv.Atoi(idS)
	if err != nil {
		logs.Error("wrong format of ID")
		c.JSON(http.StatusOK,gin.H{
			"message":"ID参数格式不正确",
			"errorCode":-1,
		})
		return
	}
	result := dal.DeleteCase(id)
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
	id,ok := c.GetQuery("id")
	if !ok {
		logs.Error("no ID")
		c.JSON(http.StatusOK,gin.H{
			"message":"没有案例id",
			"errorCode":-1,
		})
		return
	}
	solution,ok := c.GetQuery("solution")
	if (!ok || (ok && solution == "")){
		logs.Error("no solution info")
		c.JSON(http.StatusOK,gin.H{
			"message":"没有案例id",
			"errorCode":-1,
		})
		return
	}
	var data = make(map[string]string)
	data["condition"] = "id="+id
	data["solution"] = solution
	result := dal.UpdateRejCaseofSolution(data)
	if result != null {
		c.JSON(http.StatusOK,gin.H{
			"message":reslut,
			"errorCode":-1,
		})
		return
	}
	logs.Infos("edit success")
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
		return nil,err
	}
	key := fmt.Sprint(time.Now().UnixNano()) + "_" + fileName
	logs.Info("key: " + key)
	err = tosPutClient.PutObject(context, key, int64(len(byte)), bytes.NewBuffer(byte))
	if err != nil {
		logs.Error("%s", "上传tos失败：" + err.Error())
		return nil,err
	}
	domains := tos.GetDomainsForLargeFile("TT", path)
	domain := domains[rand.Intn(len(domains)-1)]
	domain = "tosv.byted.org/obj/" + _const.TOS_BUCKET_NAME
	var returnUrl string
	returnUrl = "https://" + domain + "/" + key
	//dal.UpdateDetectTosUrl(returnUrl, taskId)
	return returnUrl, nil
}