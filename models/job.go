package models

type Job struct {
	Versioned
	Status       string  `gorm:"column:status"`
	Rate         float64 `gorm:"column:rate"`
	Title        string  `gorm:"column:title"`
	CompanyID    string  `gorm:"column:company_id"`
	ContractorID string  `gorm:"column:contractor_id"`
} 