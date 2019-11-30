package database

import (
	"code.byted.org/gopkg/gorm"
	"code.byted.org/gopkg/logs"
)

// DeleteDBRecord deletes all eligbile records in the given database table.
// Note that the parameter table should be the instance or pointer of the struct
// mapping to the table and table name will be invalid.
// If the model has a DeletedAt field, it will get a soft delete ability
// automatically, which means the record will not be permanently removed
// from the database, rather the DeletedAt' value will be set to the current time.
func DeleteDBRecord(db *gorm.DB, table interface{}, sieve map[string]interface{}) error {

	if err := db.Debug().Delete(table, sieve).Error; err != nil {
		logs.Error("database error: %v", err)
		return err
	}

	return nil
}

// InsertDBRecord inserts record into database if it is valid.
// Note that the parameter record must be the pointer of the struct mapping to
// the target table.
// If the operation succeeded, record.ID is exact the id of the inserted record.
func InsertDBRecord(db *gorm.DB, record interface{}) error {

	if err := db.Debug().Create(record).Error; err != nil {
		logs.Error("database error: %v", err)
		return err
	}

	return nil
}

// UpdateDBRecord updates the specific record if it is valid.
// Note that the parameter record must be the pointer of the struct mapping to
// the target table and the primary key of record must be valid.
// WARNING: when update with struct, GORM will only update those fields
// that with non blank value. So it is necessary to update the status of
// detect task with a new function.
func UpdateDBRecord(db *gorm.DB, record interface{}) error {

	if err := db.Debug().Model(record).Updates(record).Error; err != nil {
		logs.Error("database error: %v", err)
		return err
	}

	return nil
}
