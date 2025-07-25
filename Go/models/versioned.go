package models

import (
	"gorm.io/gorm"
)

type Versioned struct {
	ID      string `gorm:"primaryKey;column:id"`
	Version int    `gorm:"primaryKey;column:version"`
	UID     string `gorm:"uniqueIndex;column:uid"`
}

func (v *Versioned) BeforeUpdate(tx *gorm.DB) (err error) {
	var maxVersion int
	err = tx.Model(v).Where("id = ?", v.ID).Select("MAX(version)").Scan(&maxVersion).Error
	if err != nil {
		return err
	}
	v.Version = maxVersion + 1
	// Instead of updating, create a new record
	tx.Statement.Model = v
	err = tx.Create(v).Error
	if err != nil {
		return err
	}
	// Cancel the update
	return gorm.ErrInvalidData
} 