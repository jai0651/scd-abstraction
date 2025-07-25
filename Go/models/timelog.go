package models

import "time"

type Timelog struct {
	Versioned
	Duration   float64   `gorm:"column:duration"`
	TimeStart  time.Time `gorm:"column:time_start"`
	TimeEnd    time.Time `gorm:"column:time_end"`
	Type       string    `gorm:"column:type"`
	JobUID     string    `gorm:"column:job_uid"`
} 