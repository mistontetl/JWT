package models

import "time"

type Billing_requests struct {
	//Model!!!
	ID     int64
	Status int // Defualt PENDING (1) !!
}

type BillingRequest struct {
	RequestToken    string  `gorm:"column:request_token;type:uuid;default:uuid_generate_v4();primaryKey"`
	UserInputTicket string  `gorm:"column:source_ticket_id;size:120;not null"`
	SourceSystem    string  `gorm:"column:source_system;size:100;not null"`
	TkID            *int64  `gorm:"column:tk_id"`
	TotalAmount     float64 `gorm:"column:total_amount;type:numeric(19,6);default:0.0"`
	RFC             string  `gorm:"column:rfc;size:13;not null"`
	StatusID        uint    `gorm:"column:status_id;not null;default:1"`
	//	Status          uint `gorm:"foreignKey:StatusID"`
	Error        *string   `gorm:"column:error;type:text"`
	BrRetryCount int       `gorm:"column:br_retry_count;default:0"`
	CreatedAt    time.Time `gorm:"column:created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at"`
}

func (BillingRequest) TableName() string {
	return "rsg_billing_requests"
}

/*
type BillingRequest struct {
	RequestToken    string        `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserInputTicket string        `gorm:"size:50;not null"`
	TotalAmount     float64       `gorm:"not null"`
	ClientEmail     string        `gorm:"size:150;not null"`
	RFC             string        `gorm:"size:13;not null"`
	StatusID        uint          `gorm:"not null"`
	Status          BillingStatus `gorm:"foreignKey:StatusID"`
	Error           *string       `gorm:"type:text"`
	BrRetryCount    int           `gorm:"column:br_retry_count"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
*/
