package detect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/clientQA/itc-server/database/dal"
	"code.byted.org/clientQA/itc-server/settings"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

// UploadFile now only supports apk/ipa/aab format.
func UploadFile(c *gin.Context) {

	//解析上传文件
	filepath, filename, ok := getFilesFromRequest(c, "uploadFile", true)
	if !ok {
		logs.Error("Failed to parse file %v", filename)
		return
	}
	mFilepath, mFilename, ok2 := getFilesFromRequest(c, "mappingFile", false)
	if !ok2 {
		logs.Error("Failed to parse file %v", filename)
		return
	}
	if mFilepath != "" && !strings.HasSuffix(mFilename, ".txt") {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("invalid mapping file: %v", mFilename))
		return
	}

	checkItem := c.DefaultPostForm("checkItem", "")
	if checkItem == "" {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("invalid detect tool type(%v)", checkItem))
		return
	}

	var task dal.DetectStruct
	if err := checkUploadParameter(&task, c, filename, mFilename); err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("invalid parameter: %v", err))
		return
	}
	task.Status = TaskStatusRunning
	if err := database.InsertDBRecord(database.DB(), &task); err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("create detect task failed: %v", err))
		return
	}
	go func() {
		msgHeader := fmt.Sprintf("task id: %v", task.ID)
		logs.Info("%s start to call detect tool", msgHeader)
		bodyBuffer := &bytes.Buffer{}
		bodyWriter := multipart.NewWriter(bodyBuffer)
		bodyWriter.WriteField("recipients", task.Creator)
		bodyWriter.WriteField("callback", settings.Get().Detect.ToolCallbackURL)
		bodyWriter.WriteField("taskID", fmt.Sprint(task.ID))
		bodyWriter.WriteField("toolIds", checkItem)
		if err := createFormFile(bodyWriter, "file", filepath, &msgHeader); err != nil {
			logs.Error("%s create form file failed: %v", msgHeader, err)
			return
		}
		if mFilepath != "" {
			if err := createFormFile(bodyWriter, "mFile", mFilepath, &msgHeader); err != nil {
				logs.Error("%s create form file failed: %v", msgHeader, err)
				return
			}
		}
		if err := bodyWriter.Close(); err != nil {
			logs.Error("%s Writer close error: %v", msgHeader, err)
			return
		}
		logs.Notice("%s Length of buffer: %vB", msgHeader, bodyBuffer.Len())
		var url string
		switch task.Platform {
		case platformAndorid:
			url = settings.Get().Detect.ToolURL + "/apk_post/v2"
		case platformiOS:
			url = settings.Get().Detect.ToolURL + "/ipa_post/v2"
		default:
			logs.Error("%s invalid platform (%v)", msgHeader, task.Platform)
			return
		}

		client := &http.Client{
			Timeout: 600 * time.Second,
			Transport: &http.Transport{
				DisableKeepAlives:   true,
				MaxIdleConnsPerHost: 0},
		}
		response, err := client.Post(url, bodyWriter.FormDataContentType(), bodyBuffer)
		if err != nil {
			go detectTaskFail(&task, fmt.Sprintf("upload file to detect tool failed: %v", err))
			return
		}
		defer response.Body.Close()

		resBody := &bytes.Buffer{}
		if n, err := resBody.ReadFrom(response.Body); err != nil {
			logs.Error("%s read form (read: %v) failed: %v", msgHeader, n, err)
			return
		}
		logs.Info("%s response of detect tool when uploading file: %s", msgHeader, resBody.Bytes())

		data := make(map[string]interface{})
		if err := json.Unmarshal(resBody.Bytes(), &data); err != nil {
			logs.Error("%s unmarshal error: %v", msgHeader, err)
			return
		}
		if fmt.Sprint(data["success"]) != "1" {
			go detectTaskFail(&task, fmt.Sprintf("detect tool error: %v", data["msg"]))
		}
	}()

	utils.ReturnMsg(c, http.StatusOK, utils.SUCCESS, "create detect task success", map[string]interface{}{"taskId": task.ID})
}

//emptyError标识该文件必须上传，且对文件大小有要求（大于1M）
func getFilesFromRequest(c *gin.Context, fieldName string, emptyError bool) (string, string, bool) {
	file, header, _ := c.Request.FormFile(fieldName)
	if file == nil {
		if emptyError {
			errorReturn(c, fieldName+":未选择上传的文件！", -2)
			logs.Error("未选择上传的文件！")
			return "", "", false
		} else {
			return "", "", true
		}
	}
	defer file.Close()
	//logs.Error(fmt.Sprint(header.Size << 20))
	if emptyError {
		if header.Size < (1 << 20) {
			errorReturn(c, fieldName+":上传的文件有问题（文件大小异常），请排查！", -2)
			logs.Error("上传的文件有问题（文件大小异常）")
			return "", "", false
		}
	}
	filename := header.Filename
	_tmpDir := "./tmp"
	exist, err := PathExists(_tmpDir)
	if !exist {
		if err := os.Mkdir(_tmpDir, os.ModePerm); err != nil {
			logs.Error("Failed to make directory %v", _tmpDir)
			return "", "", false
		}
	}
	filepath := _tmpDir + "/" + filename
	out, err := os.Create(filepath)
	if err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("os create failed: %v", err))
		return "", "", false
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		utils.ReturnMsg(c, http.StatusOK, utils.FAILURE, fmt.Sprintf("io copy failed: %v", err))
		return "", "", false
	}

	return filepath, filename, true
}

func checkUploadParameter(task *dal.DetectStruct, c *gin.Context, packageFile string, mappingFile string) error {

	userName, exist := c.Get("username")
	if !exist {
		return fmt.Errorf("unauthorized user: %v", userName)
	}
	task.Creator = userName.(string)

	toLarker := c.DefaultPostForm("toLarker", "") // Send lark message to person
	if toLarker == "" {
		task.ToLarker = userName.(string)
	} else {
		task.ToLarker = removeDuplicate(strings.Split(userName.(string)+","+toLarker, ","))
	}
	task.ToGroup = removeDuplicate(strings.Split(c.DefaultPostForm("toLarkGroupId", ""), ",")) // Send lark message to group

	platform := c.DefaultPostForm("platform", "")
	var err error
	task.Platform, err = strconv.Atoi(platform)
	if err != nil {
		return fmt.Errorf("invalid platform (%s): %v", platform, err)
	}
	if !((task.Platform == platformAndorid) && (strings.HasSuffix(packageFile, ".apk") || strings.HasSuffix(packageFile, ".aab"))) &&
		!((task.Platform == platformiOS) && strings.HasSuffix(packageFile, ".ipa")) {
		return fmt.Errorf("platform (%v) not match with file (%v)", task.Platform, packageFile)
	}

	task.AppId = c.DefaultPostForm("appId", "")
	if task.AppId == "" {
		return fmt.Errorf("invalid app id (%v)", task.AppId)
	}

	callbackURL := c.DefaultPostForm("callBackAddr", "")
	skip := c.DefaultPostForm("skipSelfFlag", "")
	if callbackURL != "" || skip != "" {
		var extraInfo dal.ExtraStruct
		extraInfo.CallBackAddr = callbackURL
		extraInfo.SkipSelfFlag = skip != ""
		byteExtraInfo, err := json.Marshal(extraInfo)
		if err != nil {
			return fmt.Errorf("unmarshal error: %v (callbackURL: %s skipSelfFlag: %s)", err, callbackURL, skip)
		}
		task.ExtraInfo = string(byteExtraInfo)
	}

	return nil
}

func createFormFile(w *multipart.Writer, fieldname string, filename string, msgHeader *string) error {

	fileWriterM, err := w.CreateFormFile(fieldname, filename)
	if err != nil {
		logs.Error("%s create form file error: %v", *msgHeader, err)
		return err
	}

	mFileHeader, err := os.Open(filename)
	if err != nil {
		logs.Error("%s io open failed: %v", *msgHeader, err)
		return err
	}
	defer mFileHeader.Close()

	if written, err := io.Copy(fileWriterM, mFileHeader); err != nil {
		logs.Error("%s io copy error: %v (written: %v)", *msgHeader, err, written)
		return err
	}

	// Remove temporary file
	go func() {
		if err := os.Remove(filename); err != nil {
			logs.Warn("%s os remove file(%s) failed: %v", *msgHeader, filename, err)
		}
	}()

	return nil
}

func detectTaskFail(task *dal.DetectStruct, detail string) {

	msgHeader := fmt.Sprintf("task id: %v", task.ID)
	logs.Error("%s %s", msgHeader, detail)
	go func() {
		task.Status = TaskStatusError
		if err := updateDetectTaskStatus(database.DB(), task); err != nil {
			logs.Warn("%s update detect task failed: %v", msgHeader, err)
		}
	}()
	go func() {
		if err := handleDetectTaskError(task, DetectServiceInfrastructureError, detail); err != nil {
			logs.Warn("%s update error information failed: %v", msgHeader, err)
		}
	}()
	go func() {
		for i := range _const.LowLarkPeople {
			utils.LarkDingOneInner(_const.LowLarkPeople[i], msgHeader+detail)
		}
	}()
}

func removeDuplicate(s []string) string {

	m := make(map[string]bool)
	for i := range s {
		if _, ok := m[s[i]]; !ok {
			m[s[i]] = false
		}
	}

	var result string
	for k := range m {
		result += k + ","
	}
	return result[:len(result)-1]
}
