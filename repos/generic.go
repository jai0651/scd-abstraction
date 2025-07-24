package repos

import (
	"gorm.io/gorm"
)

// LatestSubquery returns a subquery selecting id, MAX(version) grouped by id for the given model.
func LatestSubquery[T any](db *gorm.DB, model T) *gorm.DB {
	return db.Model(&model).
		Select("id, MAX(version) as max_version").
		Group("id")
} 