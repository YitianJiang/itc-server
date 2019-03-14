package dal

import (
	"code.byted.org/clientQA/pkg/database"
	"code.byted.org/gopkg/logs"
	"errors"
)

// 这些方法暂时还没有好的解决办法，只能生写如果设定一个父类，那么子类还是需要override

//////////////////////pkg config表操作//////////////////////

/*
通过map来获取这些config 表信息
*/
func GetPkgConfigByMap(inputs map[string]interface{}, excond string) *[]DBPkgConfigStruct {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	var pkgConfig []DBPkgConfigStruct
	if err := conn.Where(inputs).
		Where(excond).
		Find(&pkgConfig).Error; err != nil {
		//println(err.Error())
		logs.Error("%v", err)
		return nil
	}
	return &pkgConfig
}

func GetPkgConfigByCondition(condition string) *[]DBPkgConfigStruct {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	logs.Info("%s", condition)
	var pkgConfig []DBPkgConfigStruct
	if err := conn.Table(DBPkgConfigStruct{}.TableName()).Where(condition).
		Find(&pkgConfig).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	return &pkgConfig
}

func DelPkgConfigsByID(pkgconfID string) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Where("id = ?", pkgconfID).Delete(&DBPkgConfigStruct{}).Error; err != nil {
		logs.Error("%v", err)
		return err
	}
	return nil
}

func InsertPkgConfig(pkgConfig DBPkgConfigStruct) uint {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return 0
	}
	defer conn.Close()

	if err := conn.Create(&pkgConfig).Error; err != nil {
		return 0
	}
	return pkgConfig.ID
}

func UpdatePkgConfig(pkgConfig DBPkgConfigStruct) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Save(pkgConfig).Error; err != nil {
		logs.Error("%v", err)
		return err
	} else {
		return nil
	}
}

func DelPubConfigsByID(pubConfigID string) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Where("id = ?", pubConfigID).Delete(&DBPubConfigStruct{}).Error; err != nil {
		logs.Error("%v", err)
		return err
	}
	return nil
}

//////////////////////default pub config表操作//////////////////////
func GetDefaultPubConfigByMap(inputs map[string]interface{}) *[]DBDefaultPubConfigStruct {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	var defaultPubConfig []DBDefaultPubConfigStruct
	if err := conn.Where(inputs).
		Find(&defaultPubConfig).Error; err != nil {
		//println(err.Error())
		logs.Error("%v", err)
		return nil
	}
	return &defaultPubConfig
}

func GetContentDefaultPubConfig(aid int, platform string) []string {
	contents := make([]string, 0)
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	//if err := conn.Table("default_publish_config").Select("DISTINCT(content)").Where("aid = ? and platform = ?", aid, platform).
	//	Find(&contents).Error; err != nil {
	//	logs.Error("%v", err)
	//	return nil
	//} //上面这种不好用git
	if err := conn.Table("default_publish_config").Select("content").Where("aid = ? and platform = ?", aid, platform).
		Pluck("DISTINCT content", &contents).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	logs.Debug("%v", contents)
	return contents
}

func InsertDefaultPubConfig(defaultPubConfig DBDefaultPubConfigStruct) uint {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return 0
	}
	defer conn.Close()

	if err := conn.Create(&defaultPubConfig).Error; err != nil {
		return 0
	}
	return defaultPubConfig.ID
}

func DelDefaultPubConfigByMap(inputs map[string]interface{}) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	if err := conn.Where(inputs).
		Delete(&DBDefaultPubConfigStruct{}).Error; err != nil {
		logs.Error("%v", err)
		return errors.New("删除default_pub_config表失败，请手动删除")
	}
	return nil
}

//////////////////////pub config表操作//////////////////////

/*
通过map获取pub config表信息
*/
func GetPubConfigByMap(inputs map[string]interface{}) *[]DBPubConfigStruct {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	var pubConfig []DBPubConfigStruct
	if err := conn.Where(inputs).
		Find(&pubConfig).Error; err != nil {
		//println(err.Error())
		logs.Error("%v", err)
		return nil
	}
	return &pubConfig
}

func DelPubConfigsByPkgConfID(pkgconfID string) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Where("package_config_id = ?", pkgconfID).Delete(&DBPubConfigStruct{}).Error; err != nil {
		logs.Error("%v", err)
		return err
	}
	return nil
}

func InsertPubConfig(pubConfig DBPubConfigStruct) uint {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return 0
	}
	defer conn.Close()

	if err := conn.Create(&pubConfig).Error; err != nil {
		return 0
	}
	return pubConfig.ID
}

func UpdatePubConfig(pubConfig DBPubConfigStruct) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Save(&pubConfig).Error; err != nil {
		logs.Error("%v", err)
		return err
	} else {
		return nil
	}
}

//////////////////////act config表操作//////////////////////

/*
通过map获取act config表信息
*/
func GetActConfigByMap(inputs map[string]interface{}) *[]DBActConfigStruct {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	var actConfig []DBActConfigStruct
	if err := conn.Where(inputs).
		Find(&actConfig).Error; err != nil {
		//println(err.Error())
		logs.Error("%v", err)
		return nil
	}
	return &actConfig
}

func DelActConfigsByPkgConfID(pkgconfID string) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Where("package_config_id = ?", pkgconfID).Delete(&DBActConfigStruct{}).Error; err != nil {
		logs.Error("%v", err)
		return err
	}
	return nil
}

func InsertActConfig(actConfig DBActConfigStruct) uint {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return 0
	}
	defer conn.Close()

	if err := conn.Create(&actConfig).Error; err != nil {
		return 0
	}
	return actConfig.ID
}

func UpdateActConfig(actConfig DBActConfigStruct) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Save(&actConfig).Error; err != nil {
		logs.Error("%v", err)
		return err
	} else {
		return nil
	}
}

func DelActConfigsByID(actConfigID string) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Where("id = ?", actConfigID).Delete(&DBActConfigStruct{}).Error; err != nil {
		logs.Error("%v", err)
		return err
	}
	return nil
}

//////////////////////pkg表操作//////////////////////

/*
通过map获取pkg表信息
*/
func GetPkgByMap(inputs map[string]interface{}, excond string) *[]DBPkgStruct {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	var pkg []DBPkgStruct
	if err := conn.Where(inputs).
		Where(excond).
		Find(&pkg).Error; err != nil {
		//println(err.Error())
		logs.Error("%v", err)
		return nil
	}
	return &pkg
}

/*
通过limit和顺序查pub表
*/
func GetPkgByLimit(inputs map[string]interface{}, order string, limitNum int) *[]DBPkgStruct {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	var pkg []DBPkgStruct

	if err := conn.Where(inputs).
		Order(order).
		Limit(limitNum).
		Find(&pkg).Error; err != nil {
		//println(err.Error())
		logs.Error("%v", err)
		return nil
	}

	return &pkg
}

/*
通过map更新pkg，需要传入pkg及其id
*/
func UpdatePkgByMap(pkg *DBPkgStruct, inputs map[string]interface{}) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Model(pkg).
		Updates(inputs).Error; err != nil {
		logs.Error("%v", err)
		return err
	} else {
		return nil
	}
}

/*
直接更新整个pkg
*/
func UpdatePkg(pkg *DBPkgStruct) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Save(pkg).Error; err != nil {
		logs.Error("%v", err)
		return err
	} else {
		return nil
	}
}

/*
插入一条信息到pkg表中
*/
func CreatePkg(pkg DBPkgStruct) (uint, error) {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return 0, err
	}
	defer conn.Close()

	if err := conn.Create(&pkg).Error; err != nil {
		return 0, err
	}
	return pkg.ID, nil
}

func UpdatePkgStatusByFlowID(workflowId uint, status int) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Table(DBPkgStruct{}.TableName()).
		Where("workflow_id = ?", workflowId).
		Updates(map[string]interface{}{"status": status}).Error; err != nil {
		//println(err.Error())
		logs.Error("%v", err)
		return err
	}
	return nil
}

/*
通过map删除pkg表中的消息
*/
func DelPkgByMap(inputs map[string]interface{}) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	if err := conn.Where(inputs).
		Delete(&DBPkgStruct{}).Error; err != nil {
		logs.Error("%v", err)
		return errors.New("删除pkg表失败，请手动删除")
	}
	return nil
}

/*
更新pkg表中的构建号码
*/
func UpdatePkgBuildNumberByID(pkgID uint, stopUrl string, buildNumber string, isJenkinsJob bool) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if isJenkinsJob {
		stopUrl := stopUrl + "/" + buildNumber + "/" + "stop"
		if err := conn.Table(DBPkgStruct{}.TableName()).
			Where("id = ?", pkgID).
			Updates(map[string]interface{}{"jobsys_id": buildNumber, "stop_url": stopUrl}).Error; err != nil {
			//println(err.Error())
			logs.Error("%v", err)
			return err
		}
	} else {
		if err := conn.Table(DBPkgStruct{}.TableName()).
			Where("id = ?", pkgID).
			Updates(map[string]interface{}{"jobsys_id": buildNumber}).Error; err != nil {
			//println(err.Error())
			logs.Error("%v", err)
			return err
		}
	}

	return nil
}

//////////////////////pub表操作//////////////////////

/*
通过map获取pub表信息
*/
func GetPubByMap(inputs map[string]interface{}) *[]DBPubStruct {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	var pub []DBPubStruct
	if err := conn.Where(inputs).
		Find(&pub).Error; err != nil {
		//println(err.Error())
		logs.Error("%v", err)
		return nil
	}
	return &pub
}

/*
通过limit和顺序查pub表
*/
func GetPubByLimit(inputs map[string]interface{}, order string, limitNum int) *[]DBPubStruct {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	var pub []DBPubStruct

	if err := conn.Where(inputs).
		Order(order).
		Limit(limitNum).
		Find(&pub).Error; err != nil {
		//println(err.Error())
		logs.Error("%v", err)
		return nil
	}

	return &pub
}

/*
通过map更新pub，需要传入pkg及其id
*/
func UpdatePubByMap(pub *DBPubStruct, inputs map[string]interface{}) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Model(pub).
		Updates(inputs).Error; err != nil {
		logs.Error("%v", err)
		return err
	} else {
		return nil
	}
}

/*
直接更新整个pub
*/
func UpdatePub(pub *DBPubStruct) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Save(pub).Error; err != nil {
		logs.Error("%v", err)
		return err
	} else {
		return nil
	}
}

/*
通过map删除pub表中的消息
*/
func DelPubByMap(inputs map[string]interface{}) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	if err := conn.Where(inputs).
		Delete(&DBPubStruct{}).Error; err != nil {
		logs.Error("%v", err)
		return errors.New("删除pub表失败，请手动删除")
	}
	return nil
}

func UpdatePubsStatusByFlowID(workflowId uint, status int) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Table(DBPubStruct{}.TableName()).
		Where("workflow_id = ?", workflowId).
		Updates(map[string]interface{}{"status": status}).Error; err != nil {
		//println(err.Error())
		logs.Error("%v", err)
		return err
	}
	return nil
}

func UpdatePubStatusByID(pub *DBPubStruct, status int) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Model(pub).
		Select("status").
		Updates(map[string]interface{}{"status": status}).Error; err != nil {
		//println(err.Error())
		logs.Error("%v", err)
		return err
	}
	return nil
}

/*
插入pub表，以原子事务
*/
func CreatePubs(pubs []DBPubStruct) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return err
	}
	defer conn.Close()

	//tx := conn.Begin()
	//for _, v := range pubs {
	//	if err := tx.Create(&v); err != nil {
	//		tx.Rollback()
	//		logs.Error("%s", err.Error.Error())
	//		return errors.New(err.Error.Error())
	//	}
	//}
	//tx.Commit()

	//？？？为什么？？？我不能用事务？？？
	for _, v := range pubs {
		if err := conn.Create(&v).Error; err != nil {
			logs.Error("%s", err.Error())
		}
	}

	return nil
}

/*
更新pub表中的构建号码
*/
func UpdatePubBuildNumberByID(pubID uint, stopUrl string, buildNumber string, isJenkinsJob bool) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if isJenkinsJob {
		stopUrl := stopUrl + "/" + buildNumber + "/" + "stop"
		if err := conn.Table(DBPubStruct{}.TableName()).
			Where("id = ?", pubID).
			Updates(map[string]interface{}{"jobsys_id": buildNumber, "stop_url": stopUrl}).Error; err != nil {
			//println(err.Error())
			logs.Error("%v", err)
			return err
		}
	} else {
		if err := conn.Table(DBPubStruct{}.TableName()).
			Where("id = ?", pubID).
			Updates(map[string]interface{}{"jobsys_id": buildNumber}).Error; err != nil {
			//println(err.Error())
			logs.Error("%v", err)
			return err
		}
	}

	return nil
}

//////////////////////default action config表操作//////////////////////
func GetDefaultActConfigByMap(inputs map[string]interface{}) *[]DBDefaultActionConfigStruct {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	var defaultActConfig []DBDefaultActionConfigStruct
	if err := conn.Where(inputs).
		Find(&defaultActConfig).Error; err != nil {
		//println(err.Error())
		logs.Error("%v", err)
		return nil
	}
	return &defaultActConfig
}

func InsertDefaultActConfig(defaultActConfig DBDefaultActionConfigStruct) uint {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return 0
	}
	defer conn.Close()

	if err := conn.Create(&defaultActConfig).Error; err != nil {
		return 0
	}
	return defaultActConfig.ID
}

func DelDefaultActConfigByMap(inputs map[string]interface{}) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	if err := conn.Where(inputs).
		Delete(&DBDefaultActionConfigStruct{}).Error; err != nil {
		logs.Error("%v", err)
		return errors.New("删除default_action_config表失败，请手动删除")
	}
	return nil
}

//////////////////////act表操作//////////////////////

/*
通过map获取act表信息
*/
func GetActByMap(inputs map[string]interface{}) *[]DBActStruct {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	var act []DBActStruct
	if err := conn.Where(inputs).
		Find(&act).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	return &act
}

/*
通过map获取act表信息
*/
func GetActByOrder(inputs map[string]interface{}, order string) *[]DBActStruct {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	var act []DBActStruct
	if err := conn.Where(inputs).
		Order(order).
		Find(&act).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	return &act
}

/*
通过map更新pub，需要传入pkg及其id
*/
func UpdateActByMap(act *DBActStruct, inputs map[string]interface{}) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Model(act).
		Updates(inputs).Error; err != nil {
		logs.Error("%v", err)
		return err
	} else {
		return nil
	}
}

/*
直接更新整个act
*/
func UpdateAct(act *DBActStruct) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Save(act).Error; err != nil {
		logs.Error("%v", err)
		return err
	} else {
		return nil
	}
}

/*
插入act表，以原子事务
*/
func CreateActs(acts []DBActStruct) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return err
	}
	defer conn.Close()

	//tx := conn.Begin()
	//for _, v := range acts {
	//	if err := tx.Create(&v); err != nil {
	//		tx.Rollback()
	//		logs.Error("%s", err.Error.Error())
	//		return errors.New(err.Error.Error())
	//	}
	//}
	//tx.Commit()
	for _, v := range acts {
		if err := conn.Create(&v).Error; err != nil {
			logs.Error("%s", err.Error())
		}
	}

	return nil
}

/*
通过flow id来更新act的status
*/
func UpdateActsStatusByFlowID(workflowId uint, status int) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Table(DBActStruct{}.TableName()).
		Where("workflow_id = ?", workflowId).
		Updates(map[string]interface{}{"status": status}).Error; err != nil {
		//println(err.Error())
		logs.Error("%v", err)
		return err
	}
	return nil
}

/*
通过map删除act表中的消息
*/
func DelActByMap(inputs map[string]interface{}) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	if err := conn.Where(inputs).
		Delete(&DBActStruct{}).Error; err != nil {
		logs.Error("%v", err)
		return errors.New("删除act表失败，请手动删除")
	}
	return nil
}

func UpdateActStatusByID(act *DBActStruct, status int) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Model(act).
		Select("status").
		Updates(map[string]interface{}{"status": status}).Error; err != nil {
		//println(err.Error())
		logs.Error("%v", err)
		return err
	}
	return nil
}

/*
更新act表中的构建号码
*/
func UpdateActBuildNumberByID(actID uint, stopUrl string, buildNumber string, isJenkinsJob bool) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if isJenkinsJob {
		stopUrl := stopUrl + "/" + buildNumber + "/" + "stop"
		if err := conn.Table(DBActStruct{}.TableName()).
			Where("id = ?", actID).
			Updates(map[string]interface{}{"jobsys_id": buildNumber, "stop_url": stopUrl}).Error; err != nil {
			//println(err.Error())
			logs.Error("%v", err)
			return err
		}
	} else {
		if err := conn.Table(DBActStruct{}.TableName()).
			Where("id = ?", actID).
			Updates(map[string]interface{}{"jobsys_id": buildNumber}).Error; err != nil {
			//println(err.Error())
			logs.Error("%v", err)
			return err
		}
	}

	return nil
}

//////////////////////label表操作//////////////////////

/*
通过map获取label表信息
*/
func GetLabelByMap(inputs map[string]interface{}) *[]Label {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	var label []Label
	if err := conn.Where(inputs).
		Find(&label).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	return &label
}

/*
通过map更新label
*/
func UpdateLabelByMap(label *Label, inputs map[string]interface{}) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Model(label).
		Updates(inputs).Error; err != nil {
		logs.Error("%v", err)
		return err
	} else {
		return nil
	}
}

/*
直接更新整个label
*/
func UpdateLabel(label *Label) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Save(label).Error; err != nil {
		logs.Error("%v", err)
		return err
	} else {
		return nil
	}
}

/*
插入label表，以原子事务
*/
func InsertLabels(label []Label) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return err
	}
	defer conn.Close()

	//tx := conn.Begin()
	//for _, v := range acts {
	//	if err := tx.Create(&v); err != nil {
	//		tx.Rollback()
	//		logs.Error("%s", err.Error.Error())
	//		return errors.New(err.Error.Error())
	//	}
	//}
	//tx.Commit()
	for _, v := range label {
		if err := conn.Create(&v).Error; err != nil {
			logs.Error("%s", err.Error())
		}
	}

	return nil
}

/*
插入一条数据到label表
*/

func InsertLabel(label Label) uint {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Connection Failed:%v", err)
		return 0
	}
	defer conn.Close()
	if err := conn.Create(&label).Error; err != nil {
		logs.Error("%s", err.Error())
	}
	return label.ID
}

/*
通过map删除label
*/
func DelLabelByMap(inputs map[string]interface{}) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	if err := conn.Where(inputs).
		Delete(&Label{}).Error; err != nil {
		logs.Error("%v", err)
		return errors.New("删除label表失败，请手动删除")
	}
	return nil
}

/*
通过lableID删除label
*/
func DelLabelByLabelID(labelID string) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Where("id = ?", labelID).Delete(&Label{}).Error; err != nil {
		logs.Error("%v", err)
		return err
	}
	return nil
}

//////////////////////子pkg表操作//////////////////////
//子pkg主要用来给h5页面进行包下载使用，暂无其他作用...

/*
通过map获取子pkg表信息
*/
func GetSubPkgByMap(inputs map[string]interface{}) *[]SubPackage {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	var subPkg []SubPackage
	if err := conn.Where(inputs).
		Find(&subPkg).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	return &subPkg
}

/*
通过map更新subPkg
*/
func UpdateSubPkgByMap(subPkg *SubPackage, inputs map[string]interface{}) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Model(subPkg).
		Updates(inputs).Error; err != nil {
		logs.Error("%v", err)
		return err
	} else {
		return nil
	}
}

/*
直接更新整个subPkg
*/
func UpdateSubPkg(subPkg *SubPackage) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	if err := conn.Save(subPkg).Error; err != nil {
		logs.Error("%v", err)
		return err
	} else {
		return nil
	}
}

/*
插入subPkg表，以原子事务
*/
func InsertSubPkgs(subPkg []SubPackage) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return err
	}
	defer conn.Close()

	//tx := conn.Begin()
	//for _, v := range acts {
	//	if err := tx.Create(&v); err != nil {
	//		tx.Rollback()
	//		logs.Error("%s", err.Error.Error())
	//		return errors.New(err.Error.Error())
	//	}
	//}
	//tx.Commit()
	for _, v := range subPkg {
		if err := conn.Create(&v).Error; err != nil {
			logs.Error("%s", err.Error())
		}
	}

	return nil
}

//////////////////////子pkg索引表操作//////////////////////
//子pkg的索引表，主要做label下的分支统计

/*
通过map获取子pkg ind表信息
*/
func GetSubIndPkgByMap(inputs map[string]interface{}) *[]SubPackageIndex {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()

	var subPkgInd []SubPackageIndex
	if err := conn.Where(inputs).
		Find(&subPkgInd).Error; err != nil {
		logs.Error("%v", err)
		return nil
	}
	return &subPkgInd
}

/*
通过map删除子pkg ind表
*/
func DelSubIndPkgByMap(inputs map[string]interface{}) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return nil
	}
	defer conn.Close()
	if err := conn.Where(inputs).
		Delete(&SubPackageIndex{}).Error; err != nil {
		logs.Error("%v", err)
		return errors.New("删除sub pkg index表失败，请手动删除")
	}
	return nil
}

/*
插入subPkg表，以原子事务
*/
func InsertSubIndPkgs(subPkgInd []SubPackageIndex) error {
	conn, err := database.GetConnection()
	if err != nil {
		logs.Error("Get DB Conntection Failed: %v", err)
		return err
	}
	defer conn.Close()

	//tx := conn.Begin()
	//for _, v := range acts {
	//	if err := tx.Create(&v); err != nil {
	//		tx.Rollback()
	//		logs.Error("%s", err.Error.Error())
	//		return errors.New(err.Error.Error())
	//	}
	//}
	//tx.Commit()
	for _, v := range subPkgInd {
		if err := conn.Create(&v).Error; err != nil {
			logs.Error("%s", err.Error())
		}
	}

	return nil
}
