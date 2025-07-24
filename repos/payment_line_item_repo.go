package repos

import (
	"github.com/yourorg/scd-abstraction/models"
	"gorm.io/gorm"
	"time"
)

type PaymentLineItemRepo struct {
	DB *gorm.DB
}

func (r *PaymentLineItemRepo) FindLineItemsByContractorAndPeriod(contractorID string, from, to time.Time) ([]models.PaymentLineItem, error) {
	var items []models.PaymentLineItem
	subq := LatestSubquery(r.DB, models.PaymentLineItem{})
	err := r.DB.Model(&models.PaymentLineItem{}).
		Joins("JOIN timelogs ON payment_line_items.timelog_uid = timelogs.uid").
		Joins("JOIN jobs ON payment_line_items.job_uid = jobs.uid").
		Joins("JOIN (?) AS latest ON payment_line_items.id = latest.id AND payment_line_items.version = latest.max_version", subq).
		Where("jobs.contractor_id = ? AND timelogs.time_start >= ? AND timelogs.time_end <= ?", contractorID, from, to).
		Find(&items).Error
	return items, err
} 