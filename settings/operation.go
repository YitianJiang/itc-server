package settings

import (
	"encoding/json"
	"time"

	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

// Settings contains all customize settings.
type Settings struct {
	NightWatchman string `json:"night_watchman"`
	Detect        struct {
		TaskURL         string `json:"task_url"`
		ToolURL         string `json:"tool_url"`
		ToolCallbackURL string `json:"tool_callback_url"`
	} `json:"detect"`
	UploadNewDetection struct {
		APPID             string            `json:"app_id"              binding:"required"`
		APPSecret         string            `json:"app_secret"          binding:"required"`
		Groups            map[string]string `json:"groups"              binding:"required"`
		ClearRule         string            `json:"clear_rule"          binding:"required"`
		GroupNameTemplate string            `json:"group_name_template" binding:"required"`
		GroupDescription  string            `json:"group_description"   binding:"required"`
		DefaultPeople     []string          `json:"default_people"      binding:"required"`
	} `json:"upload_new_detections"`
}

var settings *Settings

type settingsTable struct {
	CreatedAt time.Time `gorm:"created_at"`
	ID        int       `gorm:"id"`
	Content   []byte    `gorm:"content"`
}

func (t settingsTable) TableName() string {
	return "settings_history"
}

// Load reads settings from table settings_history.
func Load(db *gorm.DB) error {

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

// Store writes data into table settings_history.
func Store(db *gorm.DB, s ...*Settings) (err error) {

	var t settingsTable
	if len(s) > 0 {
		if t.Content, err = json.Marshal(s[0]); err != nil {
			logs.Error("marshal error: %v", err)
			return
		}
	} else {
		if t.Content, err = json.Marshal(settings); err != nil {
			logs.Error("marshal error: %v", err)
			return
		}
	}

	if err = db.Debug().Create(&t).Error; err != nil {
		logs.Error("database error: %v", err)
		return
	}

	return nil
}

// Get returns the global handler of settings.
func Get() *Settings {

	return settings
}
