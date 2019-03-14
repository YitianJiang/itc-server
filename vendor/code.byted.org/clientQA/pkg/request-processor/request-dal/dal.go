package request_dal

import (
	"code.byted.org/clientQA/pkg/const"
	"code.byted.org/clientQA/pkg/database"
	"code.byted.org/clientQA/pkg/job-processor/dal"
	"code.byted.org/gopkg/logs"
	"errors"
	"math"
	"strings"
)

/**
查询product_info表
*/
func ProductsQueryDAL(inputs map[string]interface{}) *[]ProductInfo {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	var products []ProductInfo
	if err := conn.Where(inputs).
		Order("aid asc").
		Find(&products).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	return &products
}

/**
查询product_info表,需要传入group_id
*/
func ProductsQueryGroupDAL(groupId int, platform string) *[]ProductInfo {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	var products []ProductInfo
	if err := conn.Table("product_info").Select("*").Where("group_id = ? and platform = ?", groupId, platform).
		Order("aid asc").
		Find(&products).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	return &products
}

/**
查询user_product表
*/
func UserProductsQueryDAL(inputs map[string]interface{}) *[]UserConcernProduct {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	var userConcernProduct []UserConcernProduct
	if err := conn.Where(inputs).
		Find(&userConcernProduct).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	return &userConcernProduct
}

/*
更新user_product表，需要传入aid及其user_id
*/
func UserProductUpdateDAL(productInfoId string, userId string, userConcernProduct *UserConcernProduct, inputs map[string]interface{}) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	if err := conn.Model(userConcernProduct).Where("product_info_id = ? and user_id = ?", productInfoId, userId).
		Updates(inputs).Error; err != nil {
		logs.Error("%v", err)
		return err
	} else {
		return nil
	}
}

func GetProductConcernByMap(inputs map[string]interface{}) *[]UserConcernProduct {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	var concern []UserConcernProduct
	if err := conn.Where(inputs).
		Find(&concern).Error; err != nil {
		//println(err.Error())
		logs.Error("%v", err)
		return nil
	}
	return &concern

}

func InsertProductConcern(productConcern UserConcernProduct) uint {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return 0
	}
	defer conn.Close()

	if err := conn.Create(&productConcern).Error; err != nil {
		return 0
	}
	return productConcern.ID
}

/*
通过map删除pkg表中的消息
*/
func DelProductConcern(inputs map[string]interface{}) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	if err := conn.Where(inputs).
		Delete(&UserConcernProduct{}).Error; err != nil {
		logs.Error("%v", err)
		return errors.New("删除user_concern_product表失败，请手动删除")
	}
	return nil
}

//获取历史数据的分页数据
func PkgHistoryQueryDAL(searchModel SearchModel, levels string, pkgConfID string, status string) *[]dal.DBPkgStruct {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	var pkgHistory []dal.DBPkgStruct
	cond := GetConditionClause(searchModel)
	logs.Debug("%s", cond)
	logs.Debug("%v", levels)
	logs.Debug("%v", status)
	if levels == "" {
		levels = "0"
	}
	logs.Debug("%v", pkgConfID)
	if pkgConfID != "" {
		if err := conn.Table("package").Offset(searchModel.PageNumber*_const.PkgPubHistoryPageSize).Limit(_const.PkgPubHistoryPageSize).
			Select("*").
			Where(cond).
			Where("permission_level in ("+levels+")").
			Where("status in ("+status+")").
			Where("package_config_id = ?", pkgConfID).
			Order("id desc").
			Find(&pkgHistory).Error; err != nil {
			logs.Error("%v", err)
			return nil
		}

	} else {
		if err := conn.Table("package").Offset(searchModel.PageNumber * _const.PkgPubHistoryPageSize).Limit(_const.PkgPubHistoryPageSize).
			Select("*").
			Where(cond).
			Where("permission_level in (" + levels + ")").
			Where("status in (" + status + ")").
			//Where("package_config_id = ?", pkgConfID).
			Order("id desc").
			Find(&pkgHistory).Error; err != nil {
			logs.Error("%v", err)
			return nil
		}
	}

	return &pkgHistory
}

func GetHistoryPagesDAL(searchModel SearchModel, levels string, pkgConfID string, status string) int {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return -1
	}
	defer conn.Close()
	cond := GetConditionClause(searchModel)
	if levels == "" {
		levels = "0"
	}
	counts := make([]dal.DBPkgStruct, 0)

	if pkgConfID != "" {
		if err := conn.Model(&dal.DBPkgStruct{}).
			Where(cond).
			Where("permission_level in ("+levels+")").
			Where("status in ("+status+")").
			Where("package_config_id = ?", pkgConfID).
			Find(&counts).Error; err != nil {
			logs.Error("%v", err)
			return -1
		}
	} else {
		if err := conn.Model(&dal.DBPkgStruct{}).
			Where(cond).
			Where("permission_level in (" + levels + ")").
			Where("status in (" + status + ")").
			Find(&counts).Error; err != nil {
			logs.Error("%v", err)
			return -1
		}
	}

	return len(counts)
}

//获取指定workflow_id的publish表中的describe信息，前端展示需要使用
func PubHistoryQueryDAL(workflowId uint) []string {
	pubs := make([]dal.DBPubStruct, 0)
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	if err := conn.Table("publish").Select("*").Where("workflow_id = ? and active = 1", workflowId).
		Find(&pubs).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	describes := make([]string, 0)
	logs.Debug("%d", len(pubs))
	for _, item := range pubs {
		logs.Debug("%s", item.Describes)
		describes = append(describes, item.Describes)
	}
	logs.Debug("%s", describes)
	return describes
}

func GetConditionClause(searchModel SearchModel) string {
	cond := ""
	if searchModel.Aid != "" {
		cond += "aid=" + searchModel.Aid
	}
	if searchModel.Platform != "" {
		if cond != "" {
			cond += " and "
		}
		cond += "platform=\"" + strings.ToLower(searchModel.Platform) + "\""
	}
	if searchModel.Since != "" {
		if cond != "" {
			cond += " and "
		}
		cond += "created_at>\"" + searchModel.Since + "\""
	}
	if searchModel.Until != "" {
		if cond != "" {
			cond += " and "
		}
		cond += "created_at<\"" + searchModel.Until + "\""
	}
	if searchModel.CommitID != "" || searchModel.Branch != "" {
		if cond != "" {
			cond += " and ("
		}
		cond += "pkg_param like '%" + searchModel.CommitID + "%'"
	}
	if searchModel.PkgUrl != "" {
		if cond != "" {
			cond += " or "
		}
		cond += "job_url like '%" + searchModel.PkgUrl + "%'"
	}
	if searchModel.OperatorUser != "" {
		if cond != "" {
			cond += " or "
		}
		cond += "email_prefix like '%" + searchModel.OperatorUser + "%'"
	}
	if searchModel.Version != "" {
		if cond != "" {
			cond += " or "
		}
		cond += "outer_version like '%" + searchModel.Version + "%' or inner_version like '%" + searchModel.Version + "%' or five_version like '%" + searchModel.Version + "%' "
	}
	if searchModel.Content != "" {
		if cond != "" {
			cond += " or "
		}
		cond += "content like '%" + searchModel.Content + "%'"
	}
	if searchModel.Describe != "" {
		if cond != "" {
			cond += " or "
		}
		cond += "describes like '%" + searchModel.Describe + "%'"
	}
	if searchModel.WorkFlowId != "" {
		if cond != "" {
			cond += " or "
		}
		cond += "workflow_id=" + searchModel.WorkFlowId
	}
	if strings.Contains(cond, "(") {
		cond += ")"
	}
	return cond
}

func OnePkgPubHistoryQueryDAL(workflowID string) (*[]JobInfo, int, int) {
	//stage设定为较大的一个数字
	errorStage := math.MaxInt32
	pkgConfigID := 0
	retJobInfo := make([]JobInfo, 0)
	levelOneJobInfo := make([]JobInfo, 0) //action 返回的数据level=0 时是放在最开始，level=1是放在pkg和pub之间，需要区分开
	//retJobInfo中info的顺序为：1.action返回的数据level = 0的数据； 2.pkg中的数据；  3.action返回的数据level = 1的数据；  4.pub中的数据
	act := dal.GetActByMap(map[string]interface{}{
		"workflow_id": workflowID,
	})
	if act == nil {
		logs.Error("%v", "获取action数据出现错误...")
		return nil, errorStage, pkgConfigID
	}
	for _, item := range *act {
		curJobInfo := JobInfo{
			Describes:   item.Describes,
			Content:     item.Content,
			SendParam:   item.SendParam,
			InputParam:  item.ActionParam,
			ErrorLog:    item.ErrLog,
			ViewUrl:     item.ViewUrl,
			JobUrl:      item.JobUrl,
			StopUrl:     item.StopUrl,
			Status:      item.Status,
			ReturnParam: item.RetActionInfo,
			ID:          item.ID,
			JobType:     _const.Act_Job,
			StageName:   item.StageName,
			ConfigID:    int(item.ActionConfigId),
			Active:      item.Active,
			ErrLog:      item.ErrLog,
		}
		//这里分开的原因是，能够按照level的顺序来进行插入，不过效果不知道好不好，如果前端不需要，那么再改回来即可
		if item.Level <= _const.MaxPrePkgStage {
			curJobInfo.Stage = item.Level
			retJobInfo = append(retJobInfo, curJobInfo)
		} else {
			curJobInfo.Stage = item.Level
			levelOneJobInfo = append(levelOneJobInfo, curJobInfo)
		}

		if (curJobInfo.Status == _const.Status_fail || curJobInfo.Status == _const.Status_unknown) &&
			errorStage > item.Level {
			errorStage = item.Level
		}

	}
	pkg := dal.GetPkgByMap(map[string]interface{}{
		"workflow_id": workflowID,
	}, "")
	if pkg == nil {
		logs.Error("%v", "获取package数据出现错误...")
		return nil, errorStage, pkgConfigID
	}
	for _, item := range *pkg {
		curJobInfo := JobInfo{
			Describes:   item.Describes,
			Content:     item.Content,
			SendParam:   item.SendParam,
			InputParam:  item.PkgParam,
			ErrorLog:    item.ErrLog,
			ViewUrl:     item.ViewUrl,
			JobUrl:      item.JobUrl,
			StopUrl:     item.StopUrl,
			Status:      item.Status,
			ReturnParam: item.RetPkgInfo,
			ID:          item.ID,
			Stage:       _const.PkgStage,
			JobType:     _const.Pkg_Job,
			StageName:   _const.PkgStageName,
			ConfigID:    item.PackageConfigId,
			Active:      item.Active,
			ErrLog:      item.ErrLog,
		}
		retJobInfo = append(retJobInfo, curJobInfo)
		//pkg只有一个，直接拿item里面的pkgconfigid即可
		pkgConfigID = item.PackageConfigId
		if (curJobInfo.Status == _const.Status_fail || curJobInfo.Status == _const.Status_unknown) &&
			errorStage > _const.PkgStage {
			errorStage = _const.PkgStage
		}
	}
	retJobInfo = append(retJobInfo, levelOneJobInfo...)
	pub := dal.GetPubByMap(map[string]interface{}{
		"workflow_id": workflowID,
	})
	if pub == nil {
		logs.Error("%v", "获取publish数据出现错误...")
		return nil, errorStage, pkgConfigID
	}
	for _, item := range *pub {
		curJobInfo := JobInfo{
			Describes:   item.Describes,
			Content:     item.Content,
			SendParam:   item.SendParam,
			InputParam:  item.PublishParam,
			ErrorLog:    item.ErrLog,
			ViewUrl:     item.ViewUrl,
			JobUrl:      item.JobUrl,
			StopUrl:     item.StopUrl,
			Status:      item.Status,
			ReturnParam: item.RetPublishInfo,
			ID:          item.ID,
			Stage:       _const.PubStage,
			JobType:     _const.Pub_Job,
			StageName:   _const.PubStageName,
			ConfigID:    int(item.PublishConfigId),
			Active:      item.Active,
			ErrLog:      item.ErrLog,
		}
		retJobInfo = append(retJobInfo, curJobInfo)
		if (curJobInfo.Status == _const.Status_fail || curJobInfo.Status == _const.Status_unknown) &&
			errorStage > _const.PubStage {
			errorStage = _const.PubStage
		}
	}
	return &retJobInfo, errorStage, pkgConfigID
}

func GetSubPkgIndConditionClause(searchModel SearchModel) string {

	cond := ""
	if searchModel.CommitID != "" || searchModel.Branch != "" {
		cond += "(pkg_param like '%" + searchModel.CommitID + "%'"
	}
	if searchModel.OperatorUser != "" {
		if cond != "" {
			cond += " or "
		}
		cond += "email_prefix like '%" + searchModel.OperatorUser + "%'"
	}
	if searchModel.Version != "" {
		if cond != "" {
			cond += " or "
		}
		cond += "outer_version like '%" + searchModel.Version + "%' or inner_version like '%" + searchModel.Version + "%' or five_version like '%" + searchModel.Version + "%' "
	}
	if searchModel.Describe != "" {
		if cond != "" {
			cond += " or "
		}
		cond += "describes like '%" + searchModel.Describe + "%'"
	}
	if strings.Contains(cond, "(") {
		cond += ")"
	}
	if searchModel.LabelID != "" {
		if cond != "" {
			cond += " and "
		}
		cond += "label_id=" + searchModel.LabelID
	}
	return cond
}

//获取sub Pkg index 的分页数据
func SubPkgIndexHistory(searchModel SearchModel) *[]dal.SubPackageIndex {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	var subPkgHistory []dal.SubPackageIndex
	cond := GetSubPkgIndConditionClause(searchModel)
	logs.Debug("%s", cond)

	if err := conn.Table("sub_pakcage_index").Offset(searchModel.PageNumber * _const.PkgPubHistoryPageSize).Limit(_const.PkgPubHistoryPageSize).
		Select("*").
		Where(cond).
		//Where("package_config_id = ?", pkgConfID).
		Order("id desc").
		Find(&subPkgHistory).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	return &subPkgHistory
}

//获取sub pkg index的分页数量
func SubPkgIndexCount(searchModel SearchModel) int {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return -1
	}
	defer conn.Close()
	cond := GetSubPkgIndConditionClause(searchModel)
	logs.Info("%v", cond)
	count := 0
	if err := conn.Table("sub_pakcage_index").
		Select("*").
		Where(cond).
		Where("deleted_at is NULL").
		Count(&count).
		Error; err != nil {
		logs.Error("%v", err)
		return -1
	} else {
		return count
	}
}

func GetSubPkgConditionClause(searchModel SearchModel) string {

	cond := ""
	if searchModel.Branch != "" {
		cond += "branch='" + searchModel.Branch + "'"
	}
	if searchModel.LabelID != "" {
		if cond != "" {
			cond += " and "
		}
		cond += "label_id=" + searchModel.LabelID
	}
	if searchModel.Aid != "" {
		if cond != "" {
			cond += " and "
		}
		cond += "aid=" + searchModel.Aid
	}
	if searchModel.Platform != "" {
		if cond != "" {
			cond += " and "
		}
		cond += "platform='" + searchModel.Platform + "'"
	}
	if searchModel.OperatorUser != "" {
		if cond != "" {
			cond += " or "
		}
		cond += "email_prefix like '%" + searchModel.OperatorUser + "%'"
	}
	if searchModel.Version != "" {
		if cond != "" {
			cond += " or "
		}
		cond += "outer_version like '%" + searchModel.Version + "%' or inner_version like '%" + searchModel.Version + "%' or five_version like '%" + searchModel.Version + "%' "
	}
	if searchModel.Describe != "" {
		if cond != "" {
			cond += " or "
		}
		cond += "describes like '%" + searchModel.Describe + "%'"
	}
	if strings.Contains(cond, "(") {
		cond += ")"
	}
	return cond
}

//获取sub Pkg的分页数据
func SubPkgHistory(searchModel SearchModel) *[]dal.SubPackage {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	var subPkgHistory []dal.SubPackage
	cond := GetSubPkgConditionClause(searchModel)
	logs.Debug("%s", cond)

	if err := conn.Table("sub_pakcage").Offset(searchModel.PageNumber * _const.PkgPubHistoryPageSize).Limit(_const.PkgPubHistoryPageSize).
		Select("*").
		Where(cond).
		//Where("package_config_id = ?", pkgConfID).
		Order("id desc").
		Find(&subPkgHistory).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	return &subPkgHistory
}

//获取sub pkg的分页数量
func SubPkgCount(searchModel SearchModel) int {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return -1
	}
	defer conn.Close()
	cond := GetSubPkgConditionClause(searchModel)
	count := 0
	if err := conn.Table("sub_pakcage").
		Select("*").
		Where(cond).
		//Where("package_config_id = ?", pkgConfID).
		Count(&count).
		Error; err != nil {
		logs.Error("%v", err)
		return -1
	} else {
		return count
	}
}

//获取登录用户信息
func GetUserInfo(name string) *Struct_User {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	//使用cas来验证用户的登录信息，如果没有获得name，那么返回错误，内容为未登录
	//name := cas.Username(c.Request)
	var user Struct_User
	if name == "" {
		return nil
	}
	if err := conn.Table("users").Where("name = ?", name).First(&user).Error; err != nil {
		logs.Warn("%v", err)
		return nil
	}
	return &user
}

func CreateUser(user Struct_User) uint {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return 0
	}
	defer conn.Close()

	if err := conn.Create(&user).Error; err != nil {
		return 0
	}
	return user.ID
}

func UpdateUserInfo(full_name string, user_id uint) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return err
	}
	defer conn.Close()
	if err = conn.Model(&Struct_User{}).Where("id = ?", user_id).Update("full_name", full_name).Error; err != nil {
		return err
	}
	return nil
}

func UpdateUser(user Struct_User) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return err
	}
	defer conn.Close()
	if err := conn.Save(&user).Error; err != nil {
		logs.Error("%v", err)
		return err
	} else {
		return nil
	}
}

func GetPermissionByMap(inputs map[string]interface{}) *[]PkgLevel {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	var packageLevel []PkgLevel
	if err := conn.Where(inputs).
		Find(&packageLevel).Error; err != nil {
		//println(err.Error())
		logs.Error("%v", err)
		return nil
	}
	return &packageLevel

	//if inputs == nil {
	//
	//} else {
	//	var packageLevel []PkgLevel
	//	if err := conn.Where(inputs).
	//		Find(&packageLevel).Error; err != nil {
	//		//println(err.Error())
	//		logs.Error("%v", err)
	//		return nil
	//	}
	//	return &packageLevel
	//}
}

func InsertProduct(product ProductInfo) uint {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return 0
	}
	defer conn.Close()

	if err := conn.Create(&product).Error; err != nil {
		return 0
	}
	return product.ID
}

/*
通过map获取product信息
*/
func GetProductByMap(inputs map[string]interface{}) *[]ProductInfo {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	var info []ProductInfo
	if err := conn.Where(inputs).
		Find(&info).Error; err != nil {
		//println(err.Error())
		logs.Error("%v", err)
		return nil
	}
	return &info
}

//////////////////////获取cron job的相应信息//////////////////////

func GetCronJobByMap(inputs map[string]interface{}) *[]CronPkg {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	var c []CronPkg
	if err := conn.Where(inputs).
		Find(&c).Error; err != nil {
		//println(err.Error())
		logs.Error("%v", err)
		return nil
	}
	return &c
}

func DelCronJobByMap(inputs map[string]interface{}) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	if err := conn.Where(inputs).
		Delete(&CronPkg{}).Error; err != nil {
		logs.Error("%v", err)
		return errors.New("删除cron pkg失败，请手动删除")
	}
	return nil
}

func CreateCronJob(job CronPkg) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return err
	}
	defer conn.Close()
	if err := conn.Create(&job).Error; err != nil {
		logs.Error("%s", err.Error())
	}
	return nil
}

func UpdateCronJob(job CronPkg) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return err
	}
	defer conn.Close()
	if err := conn.Save(&job).Error; err != nil {
		logs.Error("%v", err)
		return err
	} else {
		return nil
	}
}
