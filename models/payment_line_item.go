package models

type PaymentLineItem struct {
	Versioned
	JobUID     string  `gorm:"column:job_uid"`
	TimelogUID string  `gorm:"column:timelog_uid"`
	Amount     float64 `gorm:"column:amount"`
	Status     string  `gorm:"column:status"`
} 