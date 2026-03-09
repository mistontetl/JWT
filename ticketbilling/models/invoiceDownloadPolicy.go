package models

import "time"

type InvoiceDownloadPolicy struct {
	IndpID             uint64    `gorm:"column:indp_id;primaryKey"`
	AvailabilityMonths int       `gorm:"column:availability_months;not null"`
	IsActive           bool      `gorm:"column:is_active;default:true"`
	Description        *string   `gorm:"column:description"`
	CreatedAt          time.Time `gorm:"column:created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at"`
}

func (InvoiceDownloadPolicy) TableName() string {
	return "invoice_download_policy"
}
