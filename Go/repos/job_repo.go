package repos

import (
	"github.com/yourorg/Go/models"
	"gorm.io/gorm"
)

type JobRepo struct {
	DB *gorm.DB
}

func (r *JobRepo) FindActiveJobsByCompany(companyID string) ([]models.Job, error) {
	var jobs []models.Job
	subq := LatestSubquery(r.DB, models.Job{})
	err := r.DB.Model(&models.Job{}).
		Joins("JOIN (?) AS latest ON jobs.id = latest.id AND jobs.version = latest.max_version", subq).
		Where("jobs.status = ? AND jobs.company_id = ?", "active", companyID).
		Find(&jobs).Error
	return jobs, err
}

func (r *JobRepo) FindActiveJobsByContractor(contractorID string) ([]models.Job, error) {
	var jobs []models.Job
	subq := LatestSubquery(r.DB, models.Job{})
	err := r.DB.Model(&models.Job{}).
		Joins("JOIN (?) AS latest ON jobs.id = latest.id AND jobs.version = latest.max_version", subq).
		Where("jobs.status = ? AND jobs.contractor_id = ?", "active", contractorID).
		Find(&jobs).Error
	return jobs, err
}
