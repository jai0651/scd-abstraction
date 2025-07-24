package repos

import (
	"github.com/yourorg/scd-abstraction/models"
	"gorm.io/gorm"
	"time"
)

type TimelogRepo struct {
	DB *gorm.DB
}

func (r *TimelogRepo) FindTimelogsByContractorAndPeriod(contractorID string, from, to time.Time) ([]models.Timelog, error) {
	var timelogs []models.Timelog
	subq := LatestSubquery(r.DB, models.Timelog{})
	err := r.DB.Model(&models.Timelog{}).
		Joins("JOIN jobs ON timelogs.job_uid = jobs.uid").
		Joins("JOIN (?) AS latest ON timelogs.id = latest.id AND timelogs.version = latest.max_version", subq).
		Where("jobs.contractor_id = ? AND timelogs.time_start >= ? AND timelogs.time_end <= ?", contractorID, from, to).
		Find(&timelogs).Error
	return timelogs, err
} 