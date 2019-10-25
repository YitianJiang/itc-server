package detect

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"code.byted.org/clientQA/itc-server/database"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
	"github.com/gin-gonic/gin"
)

type detectionBasic struct {
	APPID      string `json:"appid"`
	APPName    string `json:"appName"`
	APPVersion string `json:"appVersion"`
	Platform   string `json:"platform"`
	RDName     string `json:"rd_username"`
	RDEmail    string `json:"rd_email"`
	CommitID   string `json:"commitId"`
	Branch     string `json:"branch"`
}

// If the type is "敏感方法", the key is equal to className.methodName
type detectionDetail struct {
	DetectConfigID uint64         `json:"configid"`
	ClassName      string         `json:"className"`
	MethodName     string         `json:"methodName"`
	Key            string         `json:"key"`
	Description    string         `json:"desc"`
	Type           string         `json:"type"`
	GPFlag         int            `json:"gpFlag"`
	RiskLevel      int            `json:"priority"`
	Creator        string         `json:"creator"`
	CallLocations  []callLocation `json:"callLocs"`
}

type callLocation struct {
	ClassName  string `json:"class_name"`
	MethodName string `json:"method_name"`
	LineNumber string `json:"line_number"`
}

// Confirmation contains the detail information of unconfirmed sensitive
// permissions/methods/strings.
type Confirmation struct {
	detectionBasic
	Permissions      []detectionDetail `json:"new_permissions"`
	SensitiveMethods []detectionDetail `json:"new_sensiMethodCall"`
	SensitiveStrings []detectionDetail `json:"new_sensiStrCall"`
}

// NewDetection corresponds to table new_detection.
type NewDetection struct {
	ID             uint64    `gorm:"column:id"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	APPID          string    `gorm:"column:app_id"`
	APPName        string    `gorm:"column:app_name"`
	APPVersion     string    `gorm:"column:app_version"`
	Platform       string    `gorm:"column:platform"`
	RDName         string    `gorm:"column:rd_name"`
	RDEmail        string    `gorm:"column:rd_email"`
	CommitID       string    `gorm:"column:commit_id"`
	Branch         string    `gorm:"column:branch"`
	DetectConfigID uint64    `gorm:"column:detect_config_id"`
	Key            string    `gorm:"column:key_name"`
	Description    string    `gorm:"column:description"`
	Type           string    `gorm:"column:type"`
	GPFlag         int       `gorm:"column:gp_flag"`
	RiskLevel      int       `gorm:"column:risk_level"`
	Creator        string    `gorm:"column:creator"`
	CallLocations  string    `gorm:"column:call_locations"`
	Confirmed      bool      `gorm:"column:confirmed"`
	Confirmer      string    `gorm:"column:confirmer"`
}

// UploadUnconfirmedDetections writes the new detections to tables in
// database and invite the confirmor to join the specific Lark group
// in order to inform him/her to comfirm the new detections.
func UploadUnconfirmedDetections(c *gin.Context) {

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		msg := "Failed to read request body: " + err.Error()
		ReturnMsg(c, FAILURE, msg)
		return
	}

	var detections Confirmation
	if err := json.Unmarshal(body, &detections); err != nil {
		msg := "Failed to unmarshal new detections: %v"
		ReturnMsg(c, FAILURE, msg)
		return
	}

	go handleNewDetections(&detections)

	ReturnMsg(c, SUCCESS, "Receive new detections success")
	return
}

var (
	appID     string
	appSecret string
	groupName string
	message   string
)

func handleNewDetections(detections *Confirmation) {

	if err := storeNewDetections(detections); err != nil {
		logs.Error("Failed to store new detections")
		return
	}

	settings, err := getUploadNewDetectionsSettings("settings.json")
	if err != nil {
		logs.Error("Failed to get settings")
		return
	}
	appID = settings["app_id"].(string)
	appSecret = settings["app_secret"].(string)

	if len(detections.Permissions) > 0 ||
		len(detections.SensitiveMethods) > 0 ||
		len(detections.SensitiveStrings) > 0 {
		if err := informConfirmor(settings["group_name"].(string),
			detections.RDEmail,
			packMessage(detections)); err != nil {
			logs.Error("Failed to inform the confirmor %v", detections.RDEmail)
			return
		}
	} else {
		logs.Info("There are no new detections")
	}

	return
}

func getUploadNewDetectionsSettings(
	fileName string) (map[string]interface{}, error) {

	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		logs.Error("IO ReadFile failed: %v", err)
		return nil, err
	}

	result := make(map[string]interface{})
	if err := json.Unmarshal(data, &result); err != nil {
		logs.Error("Unmarshal failed: %v", err)
		return nil, err
	}

	return result["upload_new_detections"].(map[string]interface{}), nil
}

func storeNewDetections(detections *Confirmation) error {

	db, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return err
	}
	defer db.Close()

	keyMap, err := getExraDetectionKeys(db, map[string]interface{}{
		"app_id":   detections.APPID,
		"platform": detections.Platform})
	if err != nil {
		logs.Error("Failed to get unconfirmed detection keys")
		return err
	}

	// Remove detections that exist but are unconfirmed.
	removeDuplicatePermission(detections, keyMap)
	removeDuplicateSensitiveMethod(detections, keyMap)
	removeDuplicateSensitiveString(detections, keyMap)

	storeNewPermissions(db, detections)
	storeNewSensiMethods(db, detections)
	storeNewSensiStrings(db, detections)

	return nil
}

func removeDuplicatePermission(
	detections *Confirmation, keyMap map[string]bool) {

	var r []detectionDetail
	for i := range detections.Permissions {
		if _, ok := keyMap[detections.Permissions[i].Key]; !ok {
			r = append(r, detections.Permissions[i])
		}
	}
	detections.Permissions = r
}

func removeDuplicateSensitiveMethod(
	detections *Confirmation, keyMap map[string]bool) {

	var r []detectionDetail
	for i := range detections.SensitiveMethods {
		detections.SensitiveMethods[i].Key =
			detections.SensitiveMethods[i].ClassName + "." +
				detections.SensitiveMethods[i].MethodName
		if _, ok := keyMap[detections.SensitiveMethods[i].Key]; !ok {
			r = append(r, detections.SensitiveMethods[i])
		}
	}
	detections.SensitiveMethods = r
}

func removeDuplicateSensitiveString(
	detections *Confirmation, keyMap map[string]bool) {

	var r []detectionDetail
	for i := range detections.SensitiveStrings {
		if _, ok := keyMap[detections.SensitiveStrings[i].Key]; !ok {
			r = append(r, detections.SensitiveStrings[i])
		}
	}
	detections.SensitiveStrings = r
}

func getExraDetectionKeys(
	db *gorm.DB, condition map[string]interface{}) (map[string]bool, error) {

	var keys []struct {
		Key string `gorm:"column:key_name"`
	}
	if err := db.Debug().Table("new_detection").Select("key_name").
		Where(condition).Scan(&keys).Error; err != nil {
		logs.Error("Database error: %v", err)
		return nil, err
	}

	// The value of map is useless.
	result := make(map[string]bool)
	for i := range keys {
		result[keys[i].Key] = false
	}

	return result, nil
}

func storeNewPermissions(db *gorm.DB, detections *Confirmation) {

	for i := range detections.Permissions {
		// It is acceptable if one detection was damaged.
		insertDetection(db, &NewDetection{
			APPID:          detections.detectionBasic.APPID,
			APPName:        detections.detectionBasic.APPName,
			APPVersion:     detections.detectionBasic.APPVersion,
			Platform:       detections.detectionBasic.Platform,
			RDName:         detections.detectionBasic.RDName,
			RDEmail:        detections.detectionBasic.RDEmail,
			CommitID:       detections.detectionBasic.CommitID,
			Branch:         detections.detectionBasic.Branch,
			DetectConfigID: detections.Permissions[i].DetectConfigID,
			Key:            detections.Permissions[i].Key,
			Description:    detections.Permissions[i].Description,
			Type:           detections.Permissions[i].Type,
			GPFlag:         detections.Permissions[i].GPFlag,
			RiskLevel:      detections.Permissions[i].RiskLevel,
			Creator:        detections.Permissions[i].Creator,
			Confirmed:      false,
		})
	}
}

func storeNewSensiMethods(db *gorm.DB, detections *Confirmation) {

	for i := range detections.SensitiveMethods {
		// It is acceptable if one detection was damaged.
		callLocation, err := json.Marshal(
			detections.SensitiveMethods[i].CallLocations)
		if err != nil {
			logs.Error("Failed to marshal data: %v\n%v", err, detections.SensitiveMethods[i])
			continue
		}
		insertDetection(db, &NewDetection{
			APPID:          detections.detectionBasic.APPID,
			APPName:        detections.detectionBasic.APPName,
			APPVersion:     detections.detectionBasic.APPVersion,
			Platform:       detections.detectionBasic.Platform,
			RDName:         detections.detectionBasic.RDName,
			RDEmail:        detections.detectionBasic.RDEmail,
			CommitID:       detections.detectionBasic.CommitID,
			Branch:         detections.detectionBasic.Branch,
			DetectConfigID: detections.SensitiveMethods[i].DetectConfigID,
			Key:            detections.SensitiveMethods[i].Key,
			Description:    detections.SensitiveMethods[i].Description,
			Type:           detections.SensitiveMethods[i].Type,
			GPFlag:         detections.SensitiveMethods[i].GPFlag,
			RiskLevel:      detections.SensitiveMethods[i].RiskLevel,
			Creator:        detections.SensitiveMethods[i].Creator,
			CallLocations:  string(callLocation),
			Confirmed:      false,
		})
	}

}

func storeNewSensiStrings(db *gorm.DB, detections *Confirmation) {

	for i := range detections.SensitiveStrings {
		// It is acceptable if one detection was damaged.
		callLocation, err := json.Marshal(detections.SensitiveStrings[i].CallLocations)
		if err != nil {
			logs.Error("Failed to marshal data: %v\n%v", err, detections.SensitiveStrings[i])
			continue
		}
		insertDetection(db, &NewDetection{
			APPID:          detections.detectionBasic.APPID,
			APPName:        detections.detectionBasic.APPName,
			APPVersion:     detections.detectionBasic.APPVersion,
			Platform:       detections.detectionBasic.Platform,
			RDName:         detections.detectionBasic.RDName,
			RDEmail:        detections.detectionBasic.RDEmail,
			CommitID:       detections.detectionBasic.CommitID,
			Branch:         detections.detectionBasic.Branch,
			DetectConfigID: detections.SensitiveStrings[i].DetectConfigID,
			Key:            detections.SensitiveStrings[i].Key,
			Description:    detections.SensitiveStrings[i].Description,
			Type:           detections.SensitiveStrings[i].Type,
			GPFlag:         detections.SensitiveStrings[i].GPFlag,
			RiskLevel:      detections.SensitiveStrings[i].RiskLevel,
			Creator:        detections.SensitiveStrings[i].Creator,
			CallLocations:  string(callLocation),
			Confirmed:      false,
		})
	}
}

func insertDetection(db *gorm.DB, detection *NewDetection) error {

	if err := db.Debug().Create(detection).Error; err != nil {
		logs.Error("Database error: %v\n%v", err, *detection)
		return err
	}

	return nil
}

// List returns all eligible detections from table new_detection.
func List(c *gin.Context) {

	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		ReturnMsg(c, FAILURE, "Failed to read request body: "+err.Error())
		return
	}

	sieve := make(map[string]interface{})
	if err := json.Unmarshal(body, &sieve); err != nil {
		ReturnMsg(c, FAILURE, "Failed to unmarshal request body: "+err.Error())
		return
	}

	if int(sieve["page"].(float64)) <= 0 ||
		int(sieve["pageSize"].(float64)) <= 0 {
		ReturnMsg(c, FAILURE, "Invalid page or pageSize")
		return
	}

	data, total, err := getDetectionList(sieve)
	if err != nil {
		ReturnMsg(c, FAILURE, "Failed to get detection list: "+err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errorCode": SUCCESS,
		"message":   "success",
		"total":     total,
		"data":      data})
	logs.Info("Get unconfirmed detection list success")
	return
}

type detectionOutline struct {
	ID          uint64 `gorm:"column:id"          json:"id"`
	RDName      string `gorm:"column:rd_name"     json:"rd_name"`
	Key         string `gorm:"column:key_name"    json:"key_name"`
	Description string `gorm:"column:description" json:"description"`
	Type        string `gorm:"column:type"        json:"type"`
	RiskLevel   int    `gorm:"column:risk_level"  json:"risk_level"`
	Creator     string `gorm:"column:creator"     json:"creator"`
}

func getDetectionList(sieve map[string]interface{}) (
	[]detectionOutline, int, error) {

	page := int(sieve["page"].(float64))
	pageSize := int(sieve["pageSize"].(float64))
	delete(sieve, "page")
	delete(sieve, "pageSize")

	data, err := getDetectionOutline(database.DB(), sieve)
	if err != nil {
		return nil, 0, err
	}

	if len(data) <= 0 {
		logs.Warn("Cannot find any matched detection, sieve: %v", sieve)
		return nil, len(data), nil
	}

	return getFinalDetectionList(data, page, pageSize)
}

func getFinalDetectionList(data []detectionOutline, page int, pageSize int) (
	[]detectionOutline, int, error) {

	pages := len(data)/pageSize + 1
	if pages < 0 || page > pages {
		return nil, 0, errors.New("Invalid page")
	}

	var result []detectionOutline
	if page == pages {
		// Last page
		for i := (page - 1) * pageSize; i < len(data); i++ {
			result = append(result, data[i])
		}
	} else {
		for i := 0; i < pageSize; i++ {
			result = append(result, data[(page-1)*pageSize+i])
		}
	}

	return result, len(data), nil
}

func getDetectionOutline(db *gorm.DB, sieve map[string]interface{}) (
	[]detectionOutline, error) {

	var result []detectionOutline
	if err := db.Debug().Table("new_detection").
		Where(sieve).Find(&result).Error; err != nil {
		logs.Error("Failed to retrieve detection outline: %v", err)
		return nil, err
	}

	return result, nil
}

// UnconfirmedDetail returns the detail of the specific detection
// from table new_detection.
func UnconfirmedDetail(c *gin.Context) {

	id, err := getID(c)
	if err != nil {
		ReturnMsg(c, FAILURE, err.Error())
		return
	}

	result, err := getDetectionDetail(id)
	if err != nil {
		ReturnMsg(c, FAILURE, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"errorCode": SUCCESS,
		"message":   "success",
		"data":      result})
	return
}

// id for table new_detection.
func getID(c *gin.Context) (uint64, error) {

	id, exist := c.GetQuery("id")
	if !exist {
		return 0, errors.New("Miss id")
	}

	detectionID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return 0, errors.New("Parse id error: " + err.Error())
	}

	return detectionID, nil
}

func getDetectionDetail(id uint64) (map[string]interface{}, error) {

	db, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil, err
	}
	defer db.Close()

	data, err := retrieveSingleDetection(db, map[string]interface{}{
		"id": id})
	if err != nil {
		logs.Error("Failed to retrieve detection")
		return nil, err
	}

	var location []callLocation
	if data.Type != "权限" {
		if err := json.Unmarshal(
			[]byte(data.CallLocations), &location); err != nil {
			logs.Error("Unmarshal error: %v", err)
			return nil, err
		}
	}

	result := map[string]interface{}{
		"created_at":     data.CreatedAt,
		"id":             data.ID,
		"key":            data.Key,
		"risk_level":     data.RiskLevel,
		"type":           data.Type,
		"description":    data.Description,
		"platform":       data.Platform,
		"rd_name":        data.RDName,
		"rd_email":       data.RDEmail,
		"creator":        data.Creator,
		"call_locations": location,
		"app_version":    data.APPVersion,
	}

	return result, nil
}

// Confirm set the specific detection's the value of confirmed TRUE.
func Confirm(c *gin.Context) {

	userName, exist := c.Get("username")
	if !exist {
		ReturnMsg(c, FAILURE, fmt.Sprintf("Invalid user: %v", userName))
		return
	}

	id, err := getID(c)
	if err != nil {
		ReturnMsg(c, FAILURE, err.Error())
		return
	}

	if err := confirmDetection(id, userName.(string)); err != nil {
		ReturnMsg(c, FAILURE, "Confirm failed: "+err.Error())
		return
	}

	ReturnMsg(c, SUCCESS, "success")
	return
}

func confirmDetection(id uint64, userName string) error {

	db, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return err
	}
	defer db.Close()

	if _, err := retrieveSingleDetection(db, map[string]interface{}{
		"id": id,
	}); err != nil {
		return err
	}

	if err := updateDetection(db, &NewDetection{
		ID: id, Confirmer: userName}); err != nil {
		return err
	}

	return nil
}

func retrieveSingleDetection(db *gorm.DB, condition map[string]interface{}) (
	*NewDetection, error) {

	data, err := RetrieveDetection(db, condition)
	if err != nil {
		logs.Error("Failed to retrieve detection")
		return nil, err
	}

	if len(data) <= 0 {
		logs.Error("Cannot find any matched detection")
		return nil, errors.New("Cannot find any matched detection")
	}

	return &data[0], nil
}

// RetrieveDetection returns all eligible detections from table new_detection.
func RetrieveDetection(db *gorm.DB, condition map[string]interface{}) (
	[]NewDetection, error) {

	var detections []NewDetection
	if err := db.Debug().Where(condition).
		Find(&detections).Error; err != nil {
		logs.Error("Database error: %v", err)
		return nil, err
	}

	return detections, nil
}

func updateDetection(db *gorm.DB, detection *NewDetection) error {

	if err := db.Debug().Model(detection).
		Updates(map[string]interface{}{
			"confirmed": true,
			"confirmer": detection.Confirmer,
		}).Error; err != nil {
		logs.Error("Database error: %v", err)
		return err
	}

	return nil
}

// We will send message directly to the group if the confirmor is unknown.
func informConfirmor(groupName string, userEmail string, msg string) error {

	if userEmail != "" {
		// Of course we can use non-xxxSimple functions, but using
		// use xxxSimple functions make the code more readable.
		exist, err := isUserInGroupSimple(groupName, userEmail)
		if err != nil {
			return err
		}

		if !exist {
			if err := addUserToGroupSimple(
				groupName, userEmail); err != nil {
				return err
			}
		}
	}

	if err := sendLarkMessageToGroupSimple(groupName, msg); err != nil {
		return err
	}

	return nil
}

func packMessage(detections *Confirmation) string {

	var atMsg string

	openID, _, err := GetOpenIDandUserIDSimple(detections.RDEmail)
	if err != nil {
		atMsg = "未知"
	} else {
		atMsg = fmt.Sprintf("<at open_id=\"%v\"></at>", openID)
	}

	msg := " 本次编译出现新增未确认项，请前往预审平台查看确认。\n" +
		"查看与确认地址: https://rocket.bytedance.net/rocket/itc/branchCheck?biz=" +
		detections.APPID + "\n\n" +
		"【编译信息】\n" +
		"应用名称: " + detections.APPName + "\n" +
		"研发同学: " + atMsg + "\n" +
		"Branch: " + detections.Branch + "\n" +
		"COMMIT_ID: " + detections.CommitID + "\n\n"

	if len(detections.Permissions) > 0 {
		msg += "【新增权限】\n"
		for i := range detections.Permissions {
			msg += fmt.Sprintf("%v. %v\n",
				i+1, detections.Permissions[i].Key)
		}
		msg += "\n"
	}

	if len(detections.SensitiveMethods) > 0 {
		msg += "【新增敏感方法】\n"
		for i := range detections.SensitiveMethods {
			msg += fmt.Sprintf("%v. %v\n",
				i+1, detections.SensitiveMethods[i].Key)
		}
		msg += "\n"
	}

	if len(detections.SensitiveStrings) > 0 {
		msg += "【新增敏感字符串】\n"
		for i := range detections.SensitiveStrings {
			msg += fmt.Sprintf("%v. %v\n",
				i+1, detections.SensitiveStrings[i].Key)
		}
	}

	return msg
}

func isUserInGroupSimple(groupName string, userEmail string) (bool, error) {
	token, err := getTenantAccessToken(appID, appSecret)
	if err != nil {
		logs.Error("Failed to get tenant access token: %v", err)
		return false, err
	}

	return isUserInGroup(token, groupName, userEmail)
}

func isUserInGroup(token string, groupName string, userEmail string) (bool, error) {

	memberList, err := getGroupMemberList(token, groupName)
	if err != nil {
		return false, err
	}

	openID, userID, err := getOpenIDandUserID(token, userEmail)
	if err != nil {
		return false, err
	}

	for i := range memberList {
		if memberList[i].(map[string]interface{})["open_id"].(interface{}) == openID &&
			memberList[i].(map[string]interface{})["user_id"].(interface{}) == userID {
			logs.Info("User %v was already in the group %v", userEmail, groupName)
			return true, nil
		}
	}

	logs.Info("User %v was not in the group %v", userEmail, groupName)
	return false, nil
}

// sendLarkMessageToGroupSimple assumes the default robot is in the group.
func sendLarkMessageToGroupSimple(groupName string, message string) error {
	token, err := getTenantAccessToken(appID, appSecret)
	if err != nil {
		logs.Error("Failed to get tenant access token: %v", err)
		return err
	}
	logs.Info("tenant access token: %v", token)

	groupChatID, err := getGroupChatID(token, groupName)
	if err != nil {
		return err
	}

	if err := sendLarkMessageToGroup(token, groupChatID, message); err != nil {
		return err
	}

	return nil
}

func sendLarkMessageToGroup(
	token string, groupChatID string, message string) error {

	if err := sendTEXTLarkMessage(
		token, "", "", "", "", groupChatID, message); err != nil {
		logs.Error("Failed to send text lark message")
		return err
	}

	return nil
}

// rootID is the id of specific message, optional.
// If send message to user, then we need openID/userID/userEmail.
// If send message to group, then we need groupChatID.
func sendTEXTLarkMessage(token string, rootID string,
	openID string, userID string, userEmail string,
	groupChatID string, message string) error {

	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json"}

	data, err := json.Marshal(map[string]interface{}{
		"open_id":  openID,
		"root_id":  rootID,
		"chat_id":  groupChatID,
		"user_id":  userID,
		"email":    userEmail,
		"msg_type": "text", // Fixed
		"content":  map[string]interface{}{"text": message}})
	if err != nil {
		logs.Error("Marshal failed in sendTEXTLarkMessage: %v", err)
		return err
	}

	// The URL was fixed and only used here, so hard code is ok.
	body, err := SendHTTPRequest("POST",
		"https://open.feishu.cn/open-apis/message/v4/send/",
		headers, data)

	response := make(map[string]interface{})
	if err := json.Unmarshal(body, &response); err != nil {
		logs.Error("Unmarshal failed in sendTEXTLarkMessage: %v", err)
		return err
	}
	if int(response["code"].(float64)) != 0 {
		return fmt.Errorf("code: %v, message:%v", response["code"], response["msg"])
	}

	return nil
}

// addUserToGroupSimple assumes the default robot is in the group.
func addUserToGroupSimple(groupName string, userEmail string) error {

	token, err := getTenantAccessToken(appID, appSecret)
	if err != nil {
		logs.Error("Failed to get tenant access token: %v", err)
		return err
	}

	groupChatID, err := getGroupChatID(token, groupName)
	if err != nil {
		logs.Error("Failed to get chat id for group %v", groupName)
		return err
	}

	openID, userID, err := getOpenIDandUserID(token, userEmail)
	if err != nil {
		logs.Error("Failed to get open id and user id for user %v", userEmail)
		return err
	}

	if err := addUserToGroup(
		token, groupChatID, openID, userID); err != nil {
		logs.Error("Failed to add user %v to group %v: %v", userEmail, groupName, err)
		return err
	}

	logs.Info("User %v was invited to group %v", userEmail, groupName)
	return nil
}

func addUserToGroup(
	token string, groupChatID string, openID string, userID string) error {

	headers := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json"}

	data, err := json.Marshal(map[string]interface{}{
		"chat_id":  groupChatID,
		"user_ids": []interface{}{userID},
		"open_ids": []interface{}{openID}})
	if err != nil {
		logs.Error("Marshal failed in addUserToGroup: %v", err)
		return err
	}

	// The URL was fixed and only used here, so hard code is ok.
	body, err := SendHTTPRequest("POST",
		"https://open.feishu.cn/open-apis/chat/v4/chatter/add/",
		headers, data)
	if err != nil {
		logs.Error("Failed to send http request for addUserToGroup: %v", err)
		return err
	}

	response := make(map[string]interface{})
	if err := json.Unmarshal(body, &response); err != nil {
		logs.Error("Unmarshal failed in addUserToGroup: %v", err)
		return err
	}
	if int(response["code"].(float64)) != 0 {
		return fmt.Errorf("code: %v, message:%v", response["code"], response["msg"])
	}

	logs.Info("User(open_id: %v, user_id: %v) was invited to group (chat_id: %v)", openID, userID, groupChatID)
	return nil
}

func getGroupMemberList(token string, groupName string) ([]interface{}, error) {

	groupChatID, err := getGroupChatID(token, groupName)
	if err != nil {
		return nil, err
	}

	data, err := getGroupDetail(token, groupChatID)
	return data.(map[string]interface{})["members"].([]interface{}), err
}

func getGroupDetail(token string, groupChatID string) (interface{}, error) {

	header := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json"}
	// The URL was fixed and only used here, so hard code is ok.
	body, err := SendHTTPRequest("POST",
		fmt.Sprintf("https://open.feishu.cn/open-apis/chat/v4/info?chat_id=%v", groupChatID),
		header, nil)
	if err != nil {
		logs.Error("Send HTTP request failed in getGroupDetail: %v", err)
		return nil, err
	}

	response := make(map[string]interface{})
	if err := json.Unmarshal(body, &response); err != nil {
		logs.Error("Unmarshal failed in getGroupDetail: %v", err)
		return nil, err
	}
	if int(response["code"].(float64)) != 0 {
		return nil, fmt.Errorf("error code: %v, message:%v", response["code"], response["msg"])
	}

	return response["data"].(interface{}), nil
}

func getGroupChatID(token string, groupName string) (string, error) {

	groupList, err := getGroupList(token)
	if err != nil {
		return "", err
	}

	for i := range groupList {
		if groupList[i].(map[string]interface{})["name"].(string) == groupName {
			return groupList[i].(map[string]interface{})["chat_id"].(string), nil
		}
	}

	return "", fmt.Errorf("Cannot find any matched group chat id for %v", groupName)
}

func getGroupList(token string) ([]interface{}, error) {

	header := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json"}
	// The URL was fixed and only used here, so hard code is ok.
	body, err := SendHTTPRequest("POST",
		"https://open.feishu.cn/open-apis/chat/v4/list",
		header, nil)
	if err != nil {
		logs.Error("Send HTTP request failed in getGroupList: %v", err)
		return nil, err
	}

	response := make(map[string]interface{})
	if err := json.Unmarshal(body, &response); err != nil {
		logs.Error("Unmarshal failed in getGroupList: %v", err)
		return nil, err
	}

	if int(response["code"].(float64)) != 0 {
		return nil, fmt.Errorf("error code: %v, message:%v", response["code"], response["msg"])
	}

	return response["data"].(map[string]interface{})["groups"].([]interface{}), nil
}

// GetOpenIDandUserIDSimple uses the default ITC robot to retrieve
// the information of specific user.
func GetOpenIDandUserIDSimple(userEmail string) (string, string, error) {

	token, err := getTenantAccessToken(appID, appSecret)
	if err != nil {
		logs.Error("Failed to get tenant access token: %v", err)
		return "", "", err
	}
	logs.Info("tenant access token: %v", token)

	return getOpenIDandUserID(token, userEmail)
}

// The format of user email should be xxx@bytedance.com.
func getOpenIDandUserID(
	token string, userEmail string) (string, string, error) {

	data, err := json.Marshal(map[string]interface{}{
		"email": userEmail})
	if err != nil {
		logs.Error("Marshal failed in getOpenIDandUserID: %v", err)
		return "", "", err
	}

	header := map[string]string{
		"Authorization": "Bearer " + token,
		"Content-Type":  "application/json"}

	// The URL was fixed and only used here, so hard code is ok.
	body, err := SendHTTPRequest("POST",
		"https://open.feishu.cn/open-apis/user/v4/email2id",
		header, data)
	if err != nil {
		logs.Error("Send HTTP request failed in getOpenIDandUserID: %v", err)
		return "", "", err
	}

	response := make(map[string]interface{})
	if err := json.Unmarshal(body, &response); err != nil {
		logs.Error("Unmarshal failed in getOpenIDandUserID: %v", err)
		return "", "", err
	}

	if int(response["code"].(float64)) != 0 {
		return "", "", fmt.Errorf("code: %v message: %v", response["code"], response["msg"])
	}

	return response["data"].(map[string]interface{})["open_id"].(string),
		response["data"].(map[string]interface{})["user_id"].(string),
		nil
}

func getTenantAccessToken(appID string, appSecret string) (string, error) {

	data, err := json.Marshal(map[string]interface{}{
		"app_id":     appID,
		"app_secret": appSecret})
	if err != nil {
		logs.Error("Marshal failed in getTenantAccessToken: %v", err)
		return "", err
	}

	// The URL was fixed and only used here, so hard code is ok.
	body, err := SendHTTPRequest("POST",
		"https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal/",
		nil, data)
	if err != nil {
		logs.Error("Send HTTP request failed: %v", err)
		return "", err
	}

	response := make(map[string]interface{})
	if err := json.Unmarshal(body, &response); err != nil {
		logs.Error("Unmarshal failed in getTenantAccessToken: %v", err)
		return "", err
	}

	if int(response["code"].(float64)) != 0 {
		return "", fmt.Errorf("code: %v message: %v", response["code"], response["msg"])
	}

	return response["tenant_access_token"].(string), nil
}

// SendHTTPRequest uses specific method sending data to specific URL
// via HTTP request with optional authentication.
func SendHTTPRequest(method string, url string, headers map[string]string,
	data []byte) ([]byte, error) {

	// Construct HTTP handler
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		logs.Error("Construct HTTP request failed in SendHTTPRequest: %v", err)
		return nil, err
	}

	// Set request header
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	logs.Debug("%v", req)

	// Send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Error("Send HTTP request failed in SendHTTPRequest: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read HTTP response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Error("Read content from HTTP response failed in SendHTTPRequest: %v", err)
		return nil, err
	}
	logs.Debug("%v", string(body))

	return body, err
}

// Error code
const (
	FAILURE = -1
	SUCCESS = 0
)

// ReturnMsg shows necessary information for requestor.
func ReturnMsg(c *gin.Context, code int, msg string) {

	switch code {
	case FAILURE:
		logs.Error(msg)
	case SUCCESS:
		logs.Debug(msg)
	}

	c.JSON(http.StatusOK, gin.H{
		"errorCode": code,
		"message":   msg,
	})

	return
}
