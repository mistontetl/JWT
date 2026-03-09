package models

import "time"

type InvoiceSeries struct {
	InvsID       uint64    `gorm:"column:invs_id;primaryKey"`
	SourceSystem string    `gorm:"column:source_system;not null"`
	Series       string    `gorm:"column:series;not null"`
	Description  *string   `gorm:"column:description"`
	IsActive     bool      `gorm:"column:is_active"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (InvoiceSeries) TableName() string {
	return "invoice_series"
}
