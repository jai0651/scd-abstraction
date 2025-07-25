package scd

import (
	"fmt"
	"gorm.io/gorm"
	"reflect"
)

// LatestSubquery returns a subquery that selects the latest version per id
func LatestSubquery[T any](db *gorm.DB, model T) *gorm.DB {
	return db.Model(&model).
		Select("id, MAX(version) as max_version").
		Group("id")
}

// CreateNewSCDVersion clones the latest version of an entity with a new version number
func CreateNewSCDVersion[T any](db *gorm.DB, id string, updateFn func(*T)) error {
	var latest T

	// Fetch the latest version for the given ID
	if err := db.Where("id = ?", id).Order("version DESC").First(&latest).Error; err != nil {
		return fmt.Errorf("fetching latest version failed: %w", err)
	}

	// Copy the latest version to a new instance
	newVersion := latest

	// Use reflection to find and increment the Version field
	v := reflect.ValueOf(&newVersion).Elem()
	versionField := v.FieldByName("Version")
	if versionField.IsValid() && versionField.CanSet() && versionField.Kind() == reflect.Int {
		versionField.SetInt(versionField.Int() + 1)
	} else {
		return fmt.Errorf("field 'Version' not found or not settable/int in struct")
	}

	// Apply custom changes via the callback
	updateFn(&newVersion)

	// Save the new version in the DB
	if err := db.Create(&newVersion).Error; err != nil {
		return fmt.Errorf("creating new version failed: %w", err)
	}

	return nil
}
