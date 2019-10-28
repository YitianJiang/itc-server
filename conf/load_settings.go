package conf

import (
	"encoding/json"
	"io/ioutil"

	"code.byted.org/gopkg/logs"
)

// Settings contains all customize settings.
type Settings struct {
	UploadNewDetection struct {
		APPID     string            `json:"app_id"`
		APPSecret string            `json:"app_secret"`
		GroupName string            `json:"group_name"`
		Groups    map[string]string `json:"groups"`
	} `json:"upload_new_detections"`
}

const settingsFile = "settings.json"

var settings *Settings

// LoadSettings reads settings from the specific file.
func LoadSettings() error {

	data, err := ioutil.ReadFile(settingsFile)
	if err != nil {
		logs.Error("IO ReadFile failed: %v", err)
		return err
	}

	var s Settings
	s.UploadNewDetection.Groups = make(map[string]string)
	if err := json.Unmarshal(data, &s); err != nil {
		logs.Error("Unmarshal failed: %v", err)
		return err
	}
	settings = &s

	return nil
}

// GetSettings returns the global settings pointer.
func GetSettings() *Settings {

	return settings
}
