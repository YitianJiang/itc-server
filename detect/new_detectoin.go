package detect

import (
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
	APPID      string `gorm:"column:app_id"  json:"appid"       `
	APPVersion string `json:"appVersion"  gorm:"column:app_version"`
	Platform   string `json:"platform"    gorm:"column:platform"`
	RDName     string `json:"rd_username" gorm:"column:rd_username"`
	RDEmail    string `json:"rd_email"    gorm:"column:rd_email"`
}

// If the type is "敏感方法", the key is equal to className.methodName
type detectionDetail struct {
	ClassName     string         `json:"className"`
	MethodName    string         `json:"methodName"`
	Key           string         `json:"key"      gorm:"column:key_name"`
	Description   string         `json:"desc"     gorm:"column:description"`
	Type          string         `json:"type"     gorm:"column:type"`
	RiskLevel     int            `json:"priority" gorm:"column:risk_level"`
	Creator       string         `json:"creator"  gorm:"column:creator"`
	CallLocations []callLocation `json:"callLocs"`
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
	// detectionBasic
	// detectionDetail
	ID            uint64    `gorm:"column:id"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	APPID         string    `gorm:"column:app_id"`
	APPVersion    string    `gorm:"column:app_version"`
	Platform      string    `gorm:"column:platform"`
	RDName        string    `gorm:"column:rd_name"`
	RDEmail       string    `gorm:"column:rd_email"`
	Key           string    `gorm:"column:key_name"`
	Description   string    `gorm:"column:description"`
	Type          string    `gorm:"column:type"`
	RiskLevel     int       `gorm:"column:risk_level"`
	Creator       string    `gorm:"column:creator"`
	CallLocations string    `gorm:"column:call_locations"`
	Confirmed     bool      `gorm:"column:confirmed"`
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

	// printConfirmations(&detections)

	go handleNewDetections(&detections)

	ReturnMsg(c, SUCCESS, "Receive new detections success")
	return
}

func printConfirmations(detections *Confirmation) {

	fmt.Println(">>>>>>>>>> Permissions <<<<<<<<<<")
	for _, v := range detections.Permissions {
		fmt.Printf("%v %v %v %v %v %v %v %v %v %v\n",
			detections.APPID, detections.APPVersion,
			detections.Platform, detections.RDName,
			detections.RDEmail, v.Key, v.Description,
			v.RiskLevel, v.Type, v.Creator)
	}

	fmt.Println(">>>>>>>>>> Method <<<<<<<<<<")
	for _, v := range detections.SensitiveMethods {
		fmt.Printf("%v %v %v %v %v %v %v %v %v %v %v\n",
			detections.APPID, detections.APPVersion,
			detections.Platform, detections.RDName,
			detections.RDEmail, v.Key, v.Description,
			v.RiskLevel, v.Type, v.Creator, v.CallLocations)
	}

	fmt.Println(">>>>>>>>>> Strings <<<<<<<<<<")
	for _, v := range detections.SensitiveStrings {
		fmt.Printf("%v %v %v %v %v %v %v %v %v %v %v\n",
			detections.APPID, detections.APPVersion,
			detections.Platform, detections.RDName,
			detections.RDEmail, v.Key, v.Description,
			v.RiskLevel, v.Type, v.Creator, v.CallLocations)
	}
}

func handleNewDetections(detections *Confirmation) {

	if err := storeNewDetections(detections); err != nil {
		logs.Error("Failed to store new detections")
		return
	}

	if err := informConfirmor("TODO", detections.RDEmail); err != nil {
		logs.Error("Failed to inform the confirmor")
		return
	}

	return
}

// TODO
func storeNewDetections(detections *Confirmation) error {

	// Diff with exist but unconfirmed detections

	db, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil
	}
	defer db.Close()

	keyMap, err := getUncnofirmedDetectionKeys(db, map[string]interface{}{
		"app_id":    detections.APPID,
		"platform":  detections.Platform,
		"confirmed": false})
	if err != nil {
		return err
	}

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

func getUncnofirmedDetectionKeys(
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

func storeNewPermissions(db *gorm.DB, detections *Confirmation) error {

	for i := range detections.Permissions {
		if err := insertDetection(db, &NewDetection{
			APPID:       detections.detectionBasic.APPID,
			APPVersion:  detections.detectionBasic.APPVersion,
			Platform:    detections.detectionBasic.Platform,
			RDName:      detections.detectionBasic.RDName,
			RDEmail:     detections.detectionBasic.RDEmail,
			Key:         detections.Permissions[i].Key,
			Description: detections.Permissions[i].Description,
			Type:        detections.Permissions[i].Type,
			RiskLevel:   detections.Permissions[i].RiskLevel,
			Creator:     detections.Permissions[i].Creator,
			Confirmed:   false,
		}); err != nil {
			return err
		}
	}

	return nil
}

func storeNewSensiMethods(db *gorm.DB, detections *Confirmation) error {

	for i := range detections.SensitiveMethods {
		callLocation, _ := json.Marshal(
			detections.SensitiveMethods[i].CallLocations)
		if err := insertDetection(db, &NewDetection{
			APPID:         detections.detectionBasic.APPID,
			APPVersion:    detections.detectionBasic.APPVersion,
			Platform:      detections.detectionBasic.Platform,
			RDName:        detections.detectionBasic.RDName,
			RDEmail:       detections.detectionBasic.RDEmail,
			Key:           detections.SensitiveMethods[i].Key,
			Description:   detections.SensitiveMethods[i].Description,
			Type:          detections.SensitiveMethods[i].Type,
			RiskLevel:     detections.SensitiveMethods[i].RiskLevel,
			Creator:       detections.SensitiveMethods[i].Creator,
			CallLocations: string(callLocation),
			Confirmed:     false,
		}); err != nil {
			return err
		}
	}

	return nil
}

func storeNewSensiStrings(db *gorm.DB, detections *Confirmation) error {

	for i := range detections.SensitiveStrings {
		callLocation, _ := json.Marshal(detections.SensitiveStrings[i].CallLocations)
		if err := insertDetection(db, &NewDetection{
			APPID:         detections.detectionBasic.APPID,
			APPVersion:    detections.detectionBasic.APPVersion,
			Platform:      detections.detectionBasic.Platform,
			RDName:        detections.detectionBasic.RDName,
			RDEmail:       detections.detectionBasic.RDEmail,
			Key:           detections.SensitiveStrings[i].Key,
			Description:   detections.SensitiveStrings[i].Description,
			Type:          detections.SensitiveStrings[i].Type,
			RiskLevel:     detections.SensitiveStrings[i].RiskLevel,
			Creator:       detections.SensitiveStrings[i].Creator,
			CallLocations: string(callLocation),
			Confirmed:     false,
		}); err != nil {
			return err
		}
	}

	return nil
}

func insertDetection(db *gorm.DB, detection *NewDetection) error {
	if err := db.Debug().Create(detection).Error; err != nil {
		logs.Error("Failed to create record in table new_detection: %v", err.Error())
		return err
	}

	return nil
}

// UnconfirmedList returns all unconfirmed detections from table new_detection.
func UnconfirmedList(c *gin.Context) {

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
	logs.Info("Get unconfirmed detections list success")
	return
}

type detectionOutline struct {
	ID          uint64 `gorm:"column:id"          json:"id"`
	RDName      string `gorm:"column:rd_name"     json:"rd_name"`
	Key         string `gorm:"column:key_name"    json:"key"`
	Description string `gorm:"column:description" json:"description"`
	Type        string `gorm:"column:type"        json:"type"`
	RiskLevel   int    `gorm:"column:risk_level"  json:"risk_level"`
	Creator     string `gorm:"column:creator"     json:"creator"`
}

func getDetectionList(
	sieve map[string]interface{}) ([]detectionOutline, int, error) {

	db, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil, 0, err
	}
	defer db.Close()

	page := int(sieve["page"].(float64))
	pageSize := int(sieve["pageSize"].(float64))
	delete(sieve, "page")
	delete(sieve, "pageSize")
	sieve["confirmed"] = false // Only retrieve unconfirmed detections.

	data, err := getDetectionOutline(db, sieve)
	if err != nil {
		return nil, 0, err
	}

	pages := len(data)/pageSize + 1
	if pages < 0 || page > pages ||
		(page == pages && (len(data)%pageSize == 0)) {
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
		Select("id, rd_name, key_name, description, type, risk_level, creator").
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

	data, err := retrieveDetection(db, map[string]interface{}{
		"id": id})
	if err != nil {
		return nil, err
	}

	// TODO: GET APPName using RPC
	appName := "抖音"
	result := map[string]interface{}{
		"id":          data[0].ID,
		"key":         data[0].Key,
		"risk_level":  data[0].RiskLevel,
		"type":        data[0].Type,
		"decription":  data[0].Description,
		"platform":    data[0].Platform,
		"rd_name":     data[0].RDName,
		"rd_email":    data[0].RDEmail,
		"creator":     data[0].Creator,
		"app_name":    appName,
		"app_version": data[0].APPVersion,
	}

	return result, nil
}

// Confirm set the specific detection's the value of confirmed TRUE.
// TODO: insert the confirmed detection to table...
func Confirm(c *gin.Context) {

	id, err := getID(c)
	if err != nil {
		ReturnMsg(c, FAILURE, err.Error())
		return
	}

	if err := confirmDetection(id); err != nil {
		ReturnMsg(c, FAILURE, "Confirm failed: "+err.Error())
		return
	}

	ReturnMsg(c, SUCCESS, "success")
	return
}

func confirmDetection(id uint64) error {

	db, err := database.GetDBConnection()
	if err != nil {
		logs.Error("Connect to DB failed: %v", err)
		return nil
	}
	defer db.Close()

	if _, err := retrieveDetection(db, map[string]interface{}{
		"id": id,
	}); err != nil {
		return err
	}

	if err := updateDetection(db, &NewDetection{ID: id}); err != nil {
		return err
	}

	return nil
}

func retrieveDetection(db *gorm.DB, condition map[string]interface{}) (
	[]NewDetection, error) {

	var detections []NewDetection
	if err := db.Debug().Where(condition).
		Find(&detections).Error; err != nil {
		logs.Error("Failed to retrieve detections from table new_detection: %v", err)
		return nil, err
	}

	if len(detections) <= 0 {
		logs.Error("Cannot find any matched detection")
		return nil, errors.New("Cannot find any matched detection")
	}

	return detections, nil
}

func updateDetection(db *gorm.DB, detection *NewDetection) error {

	if err := db.Debug().Model(detection).
		Update("confirmed", true).Error; err != nil {
		logs.Error("Failed to update detection in table new_detection: %v", err)
		return err
	}

	return nil
}

// TODO
// We will send message directly to the group if the confirmor is unknown.
func informConfirmor(group string, emailPrefix string) error {

	if emailPrefix == "" {
	}

	return nil
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
