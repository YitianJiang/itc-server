package detect

import (
	"time"

	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

// Item is the minimum unit of detection.
type Item struct {
	Name string
	Type *string // Type will be constant, so using address here is ok.
}

// VersionDiff contains the specific version's information
// needed be attentioned of app.
// Why is attention?
// What the struct contains needs your attention.
type VersionDiff struct {
	ID        int       `gorm:"column:id"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	APPID     string    `gorm:"column:app_id"`
	Platform  int       `gorm:"column:platform"`
	Version   string    `gorm:"column:version"`
	Attention string    `gorm:"column:attention"`
}

// Attention contains the simplified information of each permission,
// sensitive method, sensitive string (or something else in the future).
type Attention struct {
	Type        string    `json:"type"`
	Status      int       `json:"status"` // 0: unconfirmed 1: pass 2: fail
	ConfirmedAt time.Time `json:"confirmed_at"`
	Confirmer   string    `json:"confirmer"`
	Remark      string    `json:"remark"`
}

// TableName returns the table name of bind with the struct.
func (VersionDiff) TableName() string {
	return "version_diff_history"
}

func readAPPAttention(db *gorm.DB, sieve map[string]interface{}) ([]VersionDiff, error) {

	var v []VersionDiff
	if err := db.Debug().Where(sieve).Find(&v).Error; err != nil {
		logs.Error("database error: %v", err)
		return nil, err
	}

	return v, nil
}
