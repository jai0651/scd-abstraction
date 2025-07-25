package repos

import (
	"github.com/yourorg/Go/scd"
	"gorm.io/gorm"
)

// LatestSubquery returns a subquery selecting id, MAX(version) grouped by id for the given model.
func LatestSubquery[T any](db *gorm.DB, model T) *gorm.DB {
	return scd.LatestSubquery(db, model)
}

// CreateNewSCDVersion creates a new SCD version for the given id and applies the updateFn.
func CreateNewSCDVersion[T any](db *gorm.DB, id string, updateFn func(*T)) error {
	return scd.CreateNewSCDVersion(db, id, updateFn)
}
