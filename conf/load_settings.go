package conf

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	_const "code.byted.org/clientQA/itc-server/const"
	"code.byted.org/clientQA/itc-server/utils"
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

// Settings contains all customize settings.
type Settings struct {
	UploadNewDetection struct {
		APPID             string            `json:"app_id"`
		APPSecret         string            `json:"app_secret"`
		Groups            map[string]string `json:"groups"`
		ClearRule         string            `json:"clear_rule"`
		GroupNameTemplate string            `json:"group_name_template"`
		GroupDescription  string            `json:"group_description"`
		DefaultPeople     []string          `json:"default_people"`
	} `json:"upload_new_detections"`
}

const settingsFile = "settings.json"
const backupSettingsFile = settingsFile + ".bak"

var settings *Settings

type settingsTable struct {
	CreatedAt time.Time `gorm:"created_at"`
	ID        int       `gorm:"id"`
	Content   []byte    `gorm:"content"`
}

func (t settingsTable) TableName() string {

	return "settings_history"
}

// LoadSettings reads settings from table settings_history.
func LoadSettings(db *gorm.DB) error {

	// data, err := ioutil.ReadFile(settingsFile)
	// if err != nil {
	// 	logs.Error("IO ReadFile failed: %v", err)
	// 	return err
	// }

	var t settingsTable
	if err := db.Debug().Last(&t).Error; err != nil {
		logs.Error("database error: %v", err)
		return err
	}

	s := new(Settings)
	s.UploadNewDetection.Groups = make(map[string]string)
	if err := json.Unmarshal(t.Content, s); err != nil {
		logs.Error("Unmarshal failed: %v", err)
		return err
	}
	settings = s

	return nil
}

// StoreSettings writes data into table settings_history.
func StoreSettings(db *gorm.DB) (err error) {

	var t settingsTable
	if t.Content, err = json.Marshal(settings); err != nil {
		logs.Error("marshal error: %v", err)
		return
	}

	if err = db.Debug().Create(&t).Error; err != nil {
		logs.Error("database error: %v", err)
		return
	}

	// data, err := json.MarshalIndent(settings, "", "    ")
	// if err != nil {
	// 	logs.Error("marshalindent error: %v", err)
	// 	return err
	// }

	// fp, err := os.OpenFile(settingsFile, os.O_RDWR, 0755)
	// if err != nil {
	// 	logs.Error("open file error: %v", err)
	// 	return err
	// }
	// defer fp.Close()

	// _, err = fp.Write(data)
	// if err != nil {
	// 	logs.Error("write file error: %v", err)
	// 	return err
	// }

	return nil
}

// BackupSettings copys settings file to backup file.
func BackupSettings() error {

	if _, err := CopyFile(backupSettingsFile, settingsFile); err != nil {
		logs.Error("backup settings file failed: %v", err)
		return err
	}

	return nil
}

// RestoreSettings moves the backup file to settings file.
func RestoreSettings() {

	if err := os.Rename(backupSettingsFile, settingsFile); err != nil {
		logs.Error("restore settings file error: %v", err)
		for i := range _const.LowLarkPeople {
			utils.LarkDingOneInner(_const.LowLarkPeople[i],
				fmt.Sprintf("restore settings file error: %v", err))
		}
	}
}

// CopyFile equal to cp srcName dstName
func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		logs.Error("os open error: %v", err)
		return
	}
	defer src.Close()

	dst, err := os.Create(dstName)
	if err != nil {
		logs.Error("os create error: %v", err)
		return
	}
	defer dst.Close()

	return io.Copy(dst, src)
}

// GetSettings returns the global settings pointer.
func GetSettings() *Settings {

	return settings
}
